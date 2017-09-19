package model

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/heartbeatsjp/happo-agent/halib"
	"github.com/heartbeatsjp/happo-agent/util"
)

// --- Global Variables
// See http://golang.org/pkg/net/http/#Client
var tr = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
var _httpClient = &http.Client{Transport: tr}

// Proxy do http reqest to next happo-agent
func Proxy(proxyRequest halib.ProxyRequest, r render.Render) (int, string) {
	var nextHostport string
	var requestType string
	var requestJSON []byte
	var err error

	nextHostport = proxyRequest.ProxyHostPort[0]

	if len(proxyRequest.ProxyHostPort) == 1 {
		// last proxy
		requestType = proxyRequest.RequestType
		requestJSON = proxyRequest.RequestJSON
	} else {
		// more proxies
		proxyRequest.ProxyHostPort = proxyRequest.ProxyHostPort[1:]
		requestType = "proxy"
		requestJSON, _ = json.Marshal(proxyRequest) // ここではエラーは出ない(出るとしたら上位でずっこけている
	}
	nextHostdata := strings.Split(nextHostport, ":")
	nextHost := nextHostdata[0]
	nextPort := halib.DefaultAgentPort
	if len(nextHostdata) == 2 {
		nextPort, err = strconv.Atoi(nextHostdata[1])
		if err != nil {
			nextPort = halib.DefaultAgentPort
		}
	}
	respCode, response, err := postToAgent(nextHost, nextPort, requestType, requestJSON)
	if err != nil {
		var monitorResponse halib.MonitorResponse
		monitorResponse.ReturnValue = halib.MonitorUnknown
		monitorResponse.Message = err.Error()
		errJSONData, _ := json.Marshal(monitorResponse)
		response = string(errJSONData[:])
	}

	return respCode, response
}

func postToAgent(host string, port int, requestType string, jsonData []byte) (int, string, error) {
	log := util.HappoAgentLogger()
	uri := fmt.Sprintf("https://%s:%d/%s", host, port, requestType)
	log.Printf("Proxy to: %s", uri)
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := _httpClient.Do(req)
	if err != nil {
		if errTimeout, ok := err.(net.Error); ok && errTimeout.Timeout() {
			return http.StatusGatewayTimeout, "", errTimeout
		}
		if resp != nil {
			if resp.StatusCode == http.StatusInternalServerError || resp.StatusCode == 0 {
				return http.StatusServiceUnavailable, "", err
			}
			return resp.StatusCode, "", err
		}
		return http.StatusInternalServerError, "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return http.StatusInternalServerError, "", err
	}
	return resp.StatusCode, string(body[:]), nil
}

// SetProxyTimeout set timeout of _httpClient
func SetProxyTimeout(timeoutSeconds int64) {
	_httpClient.Timeout = time.Duration(timeoutSeconds) * time.Second
}
