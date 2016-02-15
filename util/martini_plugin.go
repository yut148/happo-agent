package util

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-martini/martini"
)

func ACL(allowIPs []string) martini.Handler {
	return func(res http.ResponseWriter, req *http.Request, c martini.Context) {
		raw_host, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			log.Fatalln(err.Error())
		}
		host := net.ParseIP(raw_host)
		if host == nil {
			http.Error(res, "Unable to parse remote address", http.StatusForbidden)
			return
		}

		// Bypass local IP address
		if host.Equal(net.ParseIP("127.0.0.1")) {
			return
		}

		// Validate IP Addresss
		for _, raw_ip := range allowIPs {
			ip, ip_net, err := net.ParseCIDR(raw_ip)
			if err != nil {
				ip_net = nil
				ip = net.ParseIP(raw_ip)
				if ip == nil {
					http.Error(res, fmt.Sprintf("ACL format error: %s", raw_ip), http.StatusServiceUnavailable)
					return
				}
			}
			if ip.Equal(host) {
				// OK! (Equal)
				if !Production {
					log.Printf("%s <=> %s", raw_host, raw_ip)
				}
				return
			}
			if ip_net != nil && ip_net.Contains(host) {
				// OK! (Range)
				if !Production {
					log.Printf("%s <=> %s", raw_host, raw_ip)
				}
				return
			}
		}

		http.Error(res, "Access Denied", http.StatusForbidden)
	}
}

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
