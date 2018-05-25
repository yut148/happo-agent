package model

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"
	"time"

	"os"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/go-martini/martini"
	"github.com/heartbeatsjp/happo-agent/autoscaling"
	"github.com/heartbeatsjp/happo-agent/autoscaling/awsmock"
	"github.com/heartbeatsjp/happo-agent/db"
	"github.com/heartbeatsjp/happo-agent/halib"
	"github.com/martini-contrib/binding"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	leveldbUtil "github.com/syndtr/goleveldb/leveldb/util"
)

func setup() {
	//Mock
	DB, err := leveldb.Open(storage.NewMemStorage(), nil)
	if err != nil {
		os.Exit(1)
	}
	db.DB = DB
}

func teardown() {
	iter := db.DB.NewIterator(
		leveldbUtil.BytesPrefix(
			[]byte("ag-"),
		),
		nil,
	)
	for iter.Next() {
		key := iter.Key()
		db.DB.Delete(key, nil)
	}
	iter.Release()
	db.DB.Close()
}

func TestPostToAgent1(t *testing.T) {
	const stubResponse = "OK"

	ts := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, stubResponse)
			}))
	defer ts.Close()
	re, _ := regexp.Compile("([a-z]+)://([A-Za-z0-9.]+):([0-9]+)(.*)")
	found := re.FindStringSubmatch(ts.URL)
	host := found[2]
	port, _ := strconv.Atoi(found[3])

	jsonData := []byte("{}")
	statusCode, response, err := postToAgent(host, port, "test", jsonData)
	assert.EqualValues(t, http.StatusOK, statusCode)
	assert.Contains(t, response, stubResponse)
	assert.Nil(t, err)
}

func TestPostToAgent2(t *testing.T) {
	ts := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(10 * time.Millisecond)
				fmt.Fprintln(w, "will ignore(return will be blank)")
			}))
	defer ts.Close()

	re, _ := regexp.Compile("([a-z]+)://([A-Za-z0-9.]+):([0-9]+)(.*)")
	found := re.FindStringSubmatch(ts.URL)
	host := found[2]
	port, _ := strconv.Atoi(found[3])

	timeout := _httpClient.Timeout
	_httpClient.Timeout = 1 * time.Millisecond
	statusCode, response, err := postToAgent(host, port, "test", []byte("{}"))
	_httpClient.Timeout = timeout

	assert.EqualValues(t, http.StatusGatewayTimeout, statusCode)
	assert.Contains(t, response, "")
	assert.True(t, err.(net.Error).Timeout())
}

func TestPostToAgent3(t *testing.T) {
	/*
		// FIXME cannot test err != nil and err is NOT timeout.
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTemporaryRedirect)
					w.Header().Set("Location", "the/broken:location:header/")
					fmt.Fprintln(w, "will ignore(return will be blank)")
				}))
		defer ts.Close()

		re, _ := regexp.Compile("([a-z]+)://([A-Za-z0-9.]+):([0-9]+)(.*)")
		found := re.FindStringSubmatch(ts.URL)
		host := found[2]
		port, _ := strconv.Atoi(found[3])
		status_code, response, err := postToAgent(host, port, "test", []byte("{}"))

		assert.EqualValues(t, status_code, http.StatusBadGateway)
		assert.Contains(t, response, "")
		assert.NotNil(t, err)
	*/
}

func TestPostToAgent4(t *testing.T) {
	ts := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprint(w, "error response")
			}))
	defer ts.Close()

	re, _ := regexp.Compile("([a-z]+)://([A-Za-z0-9.]+):([0-9]+)(.*)")
	found := re.FindStringSubmatch(ts.URL)
	host := found[2]
	port, _ := strconv.Atoi(found[3])
	statusCode, response, err := postToAgent(host, port, "test", []byte("{}"))

	assert.EqualValues(t, http.StatusServiceUnavailable, statusCode)
	assert.Contains(t, response, "error response")
	assert.Nil(t, err)
}

func TestProxy1(t *testing.T) {
	//monitor ok

	//bastion
	m := martini.Classic()
	m.Use(render.Renderer())
	m.Post("/proxy", binding.Json(halib.ProxyRequest{}), Proxy)

	//edge
	ts := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"return_value":0,"message":"ok"}`)
			}))
	defer ts.Close()

	re, _ := regexp.Compile("([a-z]+)://([A-Za-z0-9.]+):([0-9]+)(.*)")
	found := re.FindStringSubmatch(ts.URL)
	host := found[2]
	port, _ := strconv.Atoi(found[3])

	requestJSON := fmt.Sprintf(`{
		"proxy_hostport": ["%s:%d"],
		"request_type": "monitor",
		"request_json":
			"{\"apikey\": \"\", \"plugin_name\": \"monitor_test_plugin\", \"plugin_option\": \"0\"}"
	}`, host, port)
	reader := bytes.NewReader([]byte(requestJSON))

	req, _ := http.NewRequest("POST", "/proxy", reader)
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	lastRunned = time.Now().Unix() //avoid saveMachineState
	m.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t,
		`{"return_value":0,"message":"ok"}`,
		res.Body.String(),
	)
}

func TestProxy2(t *testing.T) {
	//gateway timeout

	//bastion
	m := martini.Classic()
	m.Use(render.Renderer())
	m.Post("/proxy", binding.Json(halib.ProxyRequest{}), Proxy)

	//edge
	ts := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(1 * time.Second)
			}))
	defer ts.Close()

	re, _ := regexp.Compile("([a-z]+)://([A-Za-z0-9.]+):([0-9]+)(.*)")
	found := re.FindStringSubmatch(ts.URL)
	host := found[2]
	port, _ := strconv.Atoi(found[3])

	requestJSON := fmt.Sprintf(`{
		"proxy_hostport": ["%s:%d"],
		"request_type": "monitor",
		"request_json":
			"{\"apikey\": \"\", \"plugin_name\": \"xxx\", \"plugin_option\": \"\"}"
	}`, host, port)
	reader := bytes.NewReader([]byte(requestJSON))

	req, _ := http.NewRequest("POST", "/proxy", reader)
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	timeout := _httpClient.Timeout
	_httpClient.Timeout = 1 * time.Millisecond

	lastRunned = time.Now().Unix() //avoid saveMachineState
	m.ServeHTTP(res, req)

	_httpClient.Timeout = timeout

	assert.Equal(t, http.StatusGatewayTimeout, res.Code)
	assert.Regexp(t,
		regexp.MustCompile(
			fmt.Sprintf(`"return_value":3,"message":"Post https://%s:%d/monitor: net/http: request canceled .*(Client.Timeout exceeded while awaiting headers)`, host, port)),
		res.Body.String(),
	)
}

func TestProxy3(t *testing.T) {
	//monitor ok (multi proxy)

	//bastion
	m := martini.Classic()
	m.Use(render.Renderer())
	m.Post("/proxy", binding.Json(halib.ProxyRequest{}), Proxy)

	//edge
	ts := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"return_value":0,"message":"ok"}`)
			}))
	defer ts.Close()

	re, _ := regexp.Compile("([a-z]+)://([A-Za-z0-9.]+):([0-9]+)(.*)")
	found := re.FindStringSubmatch(ts.URL)
	host := found[2]
	port, _ := strconv.Atoi(found[3])

	requestJSON := fmt.Sprintf(`{
		"proxy_hostport": ["%s:%d","127.0.0.1:6777"],
		"request_type": "monitor",
		"request_json":
			"{\"apikey\": \"\", \"plugin_name\": \"monitor_test_plugin\", \"plugin_option\": \"0\"}"
	}`, host, port)
	reader := bytes.NewReader([]byte(requestJSON))

	req, _ := http.NewRequest("POST", "/proxy", reader)
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	lastRunned = time.Now().Unix() //avoid saveMachineState
	m.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t,
		`{"return_value":0,"message":"ok"}`,
		res.Body.String(),
	)
}

func TestProxy4(t *testing.T) {
	//gateway timeout(multi proxy)

	//bastion
	m := martini.Classic()
	m.Use(render.Renderer())
	m.Post("/proxy", binding.Json(halib.ProxyRequest{}), Proxy)

	//edge
	ts := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(1 * time.Second)
			}))
	defer ts.Close()

	re, _ := regexp.Compile("([a-z]+)://([A-Za-z0-9.]+):([0-9]+)(.*)")
	found := re.FindStringSubmatch(ts.URL)
	host := found[2]
	port, _ := strconv.Atoi(found[3])

	requestJSON := fmt.Sprintf(`{
		"proxy_hostport": ["%s:%d","127.0.0.1:6777"],
		"request_type": "monitor",
		"request_json":
			"{\"apikey\": \"\", \"plugin_name\": \"xxx\", \"plugin_option\": \"\"}"
	}`, host, port)
	reader := bytes.NewReader([]byte(requestJSON))

	req, _ := http.NewRequest("POST", "/proxy", reader)
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	timeout := _httpClient.Timeout
	_httpClient.Timeout = 1 * time.Millisecond

	lastRunned = time.Now().Unix() //avoid saveMachineState
	m.ServeHTTP(res, req)

	_httpClient.Timeout = timeout

	assert.Equal(t, http.StatusGatewayTimeout, res.Code)
	assert.Equal(t,
		fmt.Sprintf(`{"return_value":3,"message":"Post https://%s:%d/proxy: net/http: request canceled while waiting for connection (Client.Timeout exceeded while awaiting headers)"}`, host, port),
		res.Body.String(),
	)
}

func TestProxy5(t *testing.T) {
	//dispatch autoscaling instance

	setup()
	client := &autoscaling.AWSClient{
		SvcEC2:         &awsmock.MockEC2Client{},
		SvcAutoscaling: &awsmock.MockAutoScalingClient{},
	}
	autoscaling.RefreshAutoScalingInstances(client, "dummy-prod-ag", "dummy-prod-app", 10)
	defer teardown()

	//bastion
	m := martini.Classic()
	m.Use(render.Renderer())
	m.Post("/proxy", binding.Json(halib.ProxyRequest{}), Proxy)

	//edge
	ts := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"return_value":0,"message":"ok"}`)
			}))
	ts.URL = "http://192.0.2.11:6777"
	defer ts.Close()

	alias := "dummy-prod-ag-dummy-prod-app-1"
	port := 6777

	requestJSON := fmt.Sprintf(`{
		"proxy_hostport": ["%s:%d"],
		"request_type": "monitor",
		"request_json":
			"{\"apikey\": \"\", \"plugin_name\": \"monitor_test_plugin\", \"plugin_option\": \"0\"}"
	}`, alias, port)
	reader := bytes.NewReader([]byte(requestJSON))

	req, _ := http.NewRequest("POST", "/proxy", reader)
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	m.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t,
		`{"return_value":0,"message":"ok"}`,
		res.Body.String(),
	)
}
