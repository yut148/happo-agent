package model

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/heartbeatsjp/happo-agent/lib"
)

// --- Global Variables
// See http://golang.org/pkg/net/http/#Client
var tr = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
var _httpClient = &http.Client{Transport: tr}

func Proxy(proxy_request lib.ProxyRequest, r render.Render) (int, string) {
	var next_hostport string
	var request_type string
	var request_json []byte
	var err error

	next_hostport = proxy_request.ProxyHostPort[0]

	if len(proxy_request.ProxyHostPort) == 1 {
		// last proxy
		request_type = proxy_request.RequestType
		request_json = proxy_request.RequestJSON
	} else {
		// more proxies
		proxy_request.ProxyHostPort = proxy_request.ProxyHostPort[1:]
		request_type = "proxy"
		request_json, _ = json.Marshal(proxy_request) // ここではエラーは出ない(出るとしたら上位でずっこけている
	}
	next_hostdata := strings.Split(next_hostport, ":")
	next_host := next_hostdata[0]
	next_port := lib.DEFAULT_AGENT_PORT
	if len(next_hostdata) == 2 {
		next_port, err = strconv.Atoi(next_hostdata[1])
		if err != nil {
			next_port = lib.DEFAULT_AGENT_PORT
		}
	}
	resp_code, response, err := postToAgent(next_host, next_port, request_type, request_json)
	if err != nil {
		var monitor_response lib.MonitorResponse
		monitor_response.ReturnValue = lib.MONITOR_UNKNOWN
		monitor_response.Message = err.Error()
		err_jsondata, _ := json.Marshal(monitor_response)
		response = string(err_jsondata[:])
	}

	return resp_code, response
}

func postToAgent(host string, port int, request_type string, jsonData []byte) (int, string, error) {
	uri := fmt.Sprintf("https://%s:%d/%s", host, port, request_type)
	log.Printf("Proxy to: %s", uri)
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := _httpClient.Do(req)
	if err != nil {
		if errTimeout, ok := err.(net.Error); ok && errTimeout.Timeout() {
			return http.StatusGatewayTimeout, "", errTimeout
		}
		return http.StatusBadGateway, "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return http.StatusBadGateway, "", err
	}
	return resp.StatusCode, string(body[:]), nil
}

func SetProxyTimeout(timeoutSeconds int64) {
	_httpClient.Timeout = time.Duration(timeoutSeconds) * time.Second
}
