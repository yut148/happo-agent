package util

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heartbeatsjp/happo-agent/lib"

	"github.com/go-martini/martini"
	"github.com/stretchr/testify/assert"
)

func TestACL0(t *testing.T) {
	const IP = "12.12.12.12"
	const bodyStr = "success"

	m := martini.Classic()
	m.Use(ACL([]string{"FAIL"}))

	m.Get(("/test"), func() string {
		return bodyStr
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", "192.168.0.1", lib.DEFAULT_AGENT_PORT)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, res.Code, http.StatusServiceUnavailable)
}

func TestACL1(t *testing.T) {
	const IP = "12.12.12.12"
	const bodyStr = "success"

	m := martini.Classic()
	m.Use(ACL([]string{IP}))

	m.Get(("/test"), func() string {
		return bodyStr
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", "192.168.0.1", lib.DEFAULT_AGENT_PORT)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, res.Code, http.StatusForbidden)
}

func TestACL2(t *testing.T) {
	const IP = "12.12.12.12"
	const bodyStr = "success"

	m := martini.Classic()
	m.Use(ACL([]string{IP}))

	m.Get(("/test"), func() string {
		return bodyStr
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", IP, lib.DEFAULT_AGENT_PORT)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, res.Code, http.StatusOK)
	assert.EqualValues(t, res.Body.String(), bodyStr)
}

func TestACL3(t *testing.T) {
	const IP = "12.12.12.12"
	const bodyStr = "success"

	m := martini.Classic()
	m.Use(ACL([]string{IP}))

	m.Get(("/test"), func() string {
		return bodyStr
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", "127.0.0.1", lib.DEFAULT_AGENT_PORT)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, res.Code, http.StatusOK)
	assert.EqualValues(t, res.Body.String(), bodyStr)
}

func TestACL4(t *testing.T) {
	const IP = "12.12.12.12"
	const ipScope = "12.12.12.0/24"
	const bodyStr = "success"

	m := martini.Classic()
	m.Use(ACL([]string{ipScope}))

	m.Get(("/test"), func() string {
		return bodyStr
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", IP, lib.DEFAULT_AGENT_PORT)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, res.Code, http.StatusOK)
	assert.EqualValues(t, res.Body.String(), bodyStr)
}

func TestACL5(t *testing.T) {
	const IP = "12.12.12.12"
	const ipScope = "192.168.0.0/24"
	const bodyStr = "success"

	m := martini.Classic()
	m.Use(ACL([]string{ipScope}))

	m.Get(("/test"), func() string {
		return bodyStr
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", IP, lib.DEFAULT_AGENT_PORT)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, res.Code, http.StatusForbidden)
}
