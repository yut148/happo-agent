package util

import (
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-martini/martini"
)

// ACL implements AccessControlList ability
func ACL(allowIPs []string) martini.Handler {
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

type RequestStatusManager struct {
	RequestStatuses map[int64]*RequestStatus
	sync.Mutex
}
type RequestStatus struct {
	RequestCountsMap map[string]*RequestCounts
}

type RequestCounts struct {
	Counts map[int]uint64
}

func (m *RequestStatusManager) Append(when time.Time, uri string, status int) {
	var found bool
	var count uint64
	var whenKey int64

	whenKey = int64(when.Unix() / 60 * 60)

	m.Lock()
	defer m.Unlock()

	if len(m.RequestStatuses) == 0 {
		m.RequestStatuses = make(map[int64]*RequestStatus)
	}
	_, found = m.RequestStatuses[whenKey]
	if !found {
		m.RequestStatuses[whenKey] = &RequestStatus{}
	}

	if len(m.RequestStatuses[whenKey].RequestCountsMap) == 0 {
		m.RequestStatuses[whenKey].RequestCountsMap = make(map[string]*RequestCounts)
	}
	_, found = m.RequestStatuses[whenKey].RequestCountsMap[uri]
	if !found {
		m.RequestStatuses[whenKey].RequestCountsMap[uri] = &RequestCounts{}
	}

	if len(m.RequestStatuses[whenKey].RequestCountsMap[uri].Counts) == 0 {
		m.RequestStatuses[whenKey].RequestCountsMap[uri].Counts = make(map[int]uint64)
	}
	count, found = m.RequestStatuses[whenKey].RequestCountsMap[uri].Counts[status]
	if !found {
		count = 0
		m.RequestStatuses[whenKey].RequestCountsMap[uri].Counts[status] = count
	}

	m.RequestStatuses[whenKey].RequestCountsMap[uri].Counts[status] = count + 1
}

func (m *RequestStatusManager) GarbageCollect(when time.Time) {
	m.Lock()
	defer m.Unlock()

	for t, _ := range m.RequestStatuses {
		if when.Unix()-t > 60*15 {
			delete(m.RequestStatuses, t)
		}
	}
}

func (m *RequestStatusManager) GetStatus(detail bool) map[string]map[string]map[int]int {
	m.Lock()
	defer m.Unlock()

	HappoAgentLogger().Debug(spew.Sdump(rsm.RequestStatuses))
	var latestKey int64
	var keys []int64
	for t, _ := range m.RequestStatuses {
		keys = append(keys, t)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	//TODO
	HappoAgentLogger().Debug(keys)

	/*
		{
			"1min": {
				"/": {200: 3, 403: 2},
				"/?extended": {200: 1}},
			"5min": {
				"/": {200: 20, 403: 5},
				"/?extended": {200: 3},
				"/monitor": {200: 400, 500: 3}},
		}
	*/
}

var (
	rsm = &RequestStatusManager{}
)

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
			}
		}
	}()

	// cleaner
	go func() {
		for {
			select {
			case <-time.Tick(1 * time.Minute):
				rsm.GarbageCollect(time.Now())
			}
		}
	}()

	return func(res http.ResponseWriter, req *http.Request, c martini.Context) {
		c.Next()

		rw := res.(martini.ResponseWriter)
		logChan <- RequestStatusLog{When: time.Now(), URI: req.RequestURI, Status: rw.Status()}
	}
}
