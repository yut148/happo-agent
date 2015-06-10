package util

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heartbeatsjp/happo-lib"

	"github.com/go-martini/martini"
	"github.com/stretchr/testify/assert"
)

func TestACL0(t *testing.T) {
	const IP = "12.12.12.12"
	const BODY_STR = "success"

	m := martini.Classic()
	m.Use(ACL([]string{"FAIL"}))

	m.Get(("/test"), func() string {
		return BODY_STR
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", "192.168.0.1", happo_agent.DEFAULT_AGENT_PORT)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, res.Code, http.StatusServiceUnavailable)
}

func TestACL1(t *testing.T) {
	const IP = "12.12.12.12"
	const BODY_STR = "success"

	m := martini.Classic()
	m.Use(ACL([]string{IP}))

	m.Get(("/test"), func() string {
		return BODY_STR
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", "192.168.0.1", happo_agent.DEFAULT_AGENT_PORT)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, res.Code, http.StatusForbidden)
}

func TestACL2(t *testing.T) {
	const IP = "12.12.12.12"
	const BODY_STR = "success"

	m := martini.Classic()
	m.Use(ACL([]string{IP}))

	m.Get(("/test"), func() string {
		return BODY_STR
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", IP, happo_agent.DEFAULT_AGENT_PORT)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, res.Code, http.StatusOK)
	assert.EqualValues(t, res.Body.String(), BODY_STR)
}

func TestACL3(t *testing.T) {
	const IP = "12.12.12.12"
	const BODY_STR = "success"

	m := martini.Classic()
	m.Use(ACL([]string{IP}))

	m.Get(("/test"), func() string {
		return BODY_STR
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", "127.0.0.1", happo_agent.DEFAULT_AGENT_PORT)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, res.Code, http.StatusOK)
	assert.EqualValues(t, res.Body.String(), BODY_STR)
}

func TestACL4(t *testing.T) {
	const IP = "12.12.12.12"
	const IP_SCOPE = "12.12.12.0/24"
	const BODY_STR = "success"

	m := martini.Classic()
	m.Use(ACL([]string{IP_SCOPE}))

	m.Get(("/test"), func() string {
		return BODY_STR
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", IP, happo_agent.DEFAULT_AGENT_PORT)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, res.Code, http.StatusOK)
	assert.EqualValues(t, res.Body.String(), BODY_STR)
}

func TestACL5(t *testing.T) {
	const IP = "12.12.12.12"
	const IP_SCOPE = "192.168.0.0/24"
	const BODY_STR = "success"

	m := martini.Classic()
	m.Use(ACL([]string{IP_SCOPE}))

	m.Get(("/test"), func() string {
		return BODY_STR
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", IP, happo_agent.DEFAULT_AGENT_PORT)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, res.Code, http.StatusForbidden)
}
