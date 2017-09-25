package util

import (
	"encoding/json"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-martini/martini"
	"github.com/heartbeatsjp/happo-agent/halib"
)

// ACL implements AccessControlList ability
func ACL(allowIPs []string) martini.Handler {
	HappoAgentLogger().Debug("allowed hosts:", allowIPs)
	return func(res http.ResponseWriter, req *http.Request, c martini.Context) {
		log := HappoAgentLogger()
		rawHost, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			log.Fatalln(err.Error())
		}
		host := net.ParseIP(rawHost)
		if host == nil {
			http.Error(res, "Unable to parse remote address", http.StatusForbidden)
			return
		}

		// Bypass local IP address
		if host.Equal(net.ParseIP("127.0.0.1")) {
			return
		}

		// Validate IP Addresss
		for _, rawIP := range allowIPs {
			if rawIP == "" {
				continue
			}
			ip, ipNet, err := net.ParseCIDR(rawIP)
			if err != nil {
				ipNet = nil
				ip = net.ParseIP(rawIP)
				if ip == nil {
					http.Error(res, fmt.Sprintf("ACL format error: %s", rawIP), http.StatusServiceUnavailable)
					return
				}
			}
			if ip.Equal(host) {
				// OK! (Equal)
				if !Production {
					log.Printf("%s <=> %s", rawHost, rawIP)
				}
				return
			}
			if ipNet != nil && ipNet.Contains(host) {
				// OK! (Range)
				if !Production {
					log.Printf("%s <=> %s", rawHost, rawIP)
				}
				return
			}
		}
		log.WithField("RemoteAddr", host.String()).Errorf("Access Denied")
		http.Error(res, "Access Denied", http.StatusForbidden)
	}
}

// MartiniCustomLogger implements custom logger
func MartiniCustomLogger() martini.Handler {
	return func(res http.ResponseWriter, req *http.Request, c martini.Context, martiniLog *stdlog.Logger) {
		log := HappoAgentLogger()
		start := time.Now()

		addr := req.Header.Get("X-Real-IP")
		if addr == "" {
			addr = req.Header.Get("X-Forwarded-For")
			if addr == "" {
				addr = req.RemoteAddr
			}
		}

		rw := res.(martini.ResponseWriter)
		c.Next()

		log.Printf("Aceess: %s \"%s %s\" %d %d %d\n", addr, req.Method, req.RequestURI, rw.Status(), rw.Size(), time.Since(start)/time.Millisecond)
	}
}

// RequestStatusManager manages RequestStatus
type RequestStatusManager struct {
	RequestStatus []struct {
		When   int64
		URI    string
		Counts map[int]uint64
	}
	sync.Mutex
}

// Append append log to manager
func (m *RequestStatusManager) Append(when time.Time, uri string, status int) {
	whenKey := when.Unix() // round to second

	var found bool

	m.Lock()
	defer m.Unlock()

	for _, requestStatus := range m.RequestStatus {
		if requestStatus.When != whenKey {
			continue
		}
		if requestStatus.URI != uri {
			continue
		}
		found = true
		if _, ok := requestStatus.Counts[status]; ok {
			requestStatus.Counts[status] = requestStatus.Counts[status] + 1
		} else {
			requestStatus.Counts[status] = 1
		}
		break
	}
	if !found {
		requestStatus := struct {
			When   int64
			URI    string
			Counts map[int]uint64
		}{When: whenKey, URI: uri}
		requestStatus.Counts = make(map[int]uint64)
		requestStatus.Counts[status] = 1
		m.RequestStatus = append(m.RequestStatus, requestStatus)
	}
}

// GarbageCollect runs gabage collection
func (m *RequestStatusManager) GarbageCollect(when time.Time, lifetimeMinutes int64) {
	m.Lock()
	defer m.Unlock()

	var newRequestStatus []struct {
		When   int64
		URI    string
		Counts map[int]uint64
	}
	for _, requestStatus := range m.RequestStatus {
		if when.Unix()-requestStatus.When <= lifetimeMinutes*60 {
			newRequestStatus = append(newRequestStatus, requestStatus)
		}
	}
	m.RequestStatus = newRequestStatus
}

// GetStatus returns halib.RequestStatusResponse
func (m *RequestStatusManager) GetStatus(fromWhen time.Time) halib.RequestStatusResponse {
	m.Lock()
	defer m.Unlock()

	resp := halib.RequestStatusResponse{}

	if len(m.RequestStatus) == 0 {
		//no result
		return resp
	}

	for _, requestStatus := range m.RequestStatus {
		if fromWhen.Unix()-requestStatus.When <= 60 {
			// 1 min target
			foundURI := false
			for _, respData := range resp.Last1 {
				if respData.URL == requestStatus.URI {
					foundURI = true
					for k1, v1 := range requestStatus.Counts {
						_, foundCountsKey := respData.Counts[k1]
						if foundCountsKey {
							respData.Counts[k1] = respData.Counts[k1] + v1
						} else {
							respData.Counts[k1] = v1
						}
					}
				}
			}
			if !foundURI {
				data := halib.RequestStatusData{URL: requestStatus.URI, Counts: make(map[int]uint64)}
				for k, v := range requestStatus.Counts {
					data.Counts[k] = v
				}
				resp.Last1 = append(resp.Last1, data)
			}
		}

		if fromWhen.Unix()-requestStatus.When <= 300 {
			// 5 min target
			foundURI := false
			for _, respData := range resp.Last5 {
				if respData.URL == requestStatus.URI {
					foundURI = true
					for k1, v1 := range requestStatus.Counts {
						_, foundCountsKey := respData.Counts[k1]
						if foundCountsKey {
							respData.Counts[k1] = respData.Counts[k1] + v1
						} else {
							respData.Counts[k1] = v1
						}
					}
				}
			}
			if !foundURI {
				data := halib.RequestStatusData{URL: requestStatus.URI, Counts: make(map[int]uint64)}
				for k, v := range requestStatus.Counts {
					data.Counts[k] = v
				}
				resp.Last5 = append(resp.Last5, data)
			}
		}
	}
	return resp
}

var (
	rsm = &RequestStatusManager{}
)

// RequestStatusLog is data for chan
type RequestStatusLog struct {
	When   time.Time
	URI    string
	Status int
}

// MartiniRequestStatus implements recent request status
func MartiniRequestStatus() martini.Handler {
	logChan := make(chan RequestStatusLog, 1000) //FIXME proper buffer size

	// appender
	go func() {
		for {
			select {
			case log := <-logChan:
				rsm.Append(log.When, log.URI, log.Status)
				b, _ := json.Marshal(rsm.GetStatus(time.Now()))
				HappoAgentLogger().Debug(string(b))
			}
		}
	}()

	// cleaner
	go func() {
		for {
			select {
			case <-time.Tick(1 * time.Minute):
				rsm.GarbageCollect(time.Now(), 5)
			}
		}
	}()

	return func(res http.ResponseWriter, req *http.Request, c martini.Context) {
		c.Next()

		rw := res.(martini.ResponseWriter)
		logChan <- RequestStatusLog{When: time.Now(), URI: req.RequestURI, Status: rw.Status()}
	}
}

// GetMartiniRequestStatus implements recent request status
func GetMartiniRequestStatus(fromWhen time.Time) halib.RequestStatusResponse {
	return rsm.GetStatus(fromWhen)
}
