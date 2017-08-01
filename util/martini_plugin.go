package util

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-martini/martini"
)

// ACL implements AccessControlList ability
func ACL(allowIPs []string) martini.Handler {
	return func(res http.ResponseWriter, req *http.Request, c martini.Context) {
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

		http.Error(res, "Access Denied", http.StatusForbidden)
	}
}

// Logger implements custom logger
func Logger() martini.Handler {
	return func(res http.ResponseWriter, req *http.Request, c martini.Context, martiniLog *log.Logger) {
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
