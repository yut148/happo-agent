package model

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/go-martini/martini"
	"github.com/heartbeatsjp/happo-lib"
	"github.com/martini-contrib/binding"
	"github.com/stretchr/testify/assert"
)

func TestMonitor1(t *testing.T) {
	// OK

	m := martini.Classic()
	m.Use(render.Renderer())
	m.Post("/monitor", binding.Json(happo_agent.MonitorRequest{}), Monitor)

	reader := bytes.NewReader([]byte(`{
		"apikey": "",
		"plugin_name": "monitor_test_plugin",
		"plugin_option": "0"
	}`))
	req, _ := http.NewRequest("POST", "/monitor", reader)
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	lastRunned = time.Now().Unix() //avoid saveMachineState
	m.ServeHTTP(res, req)

	assert.Equal(t, res.Code, http.StatusOK)
	assert.Equal(t,
		res.Body.String(),
		`{"return_value":0,"message":"Output of monitor_test_plugin. exit status is 0\n"}`,
	)
}

func TestMonitor2(t *testing.T) {
	// Warning

	m := martini.Classic()
	m.Use(render.Renderer())
	m.Post("/monitor", binding.Json(happo_agent.MonitorRequest{}), Monitor)

	reader := bytes.NewReader([]byte(`{
		"apikey": "",
		"plugin_name": "monitor_test_plugin",
		"plugin_option": "1"
	}`))
	req, _ := http.NewRequest("POST", "/monitor", reader)
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	lastRunned = time.Now().Unix() //avoid saveMachineState
	m.ServeHTTP(res, req)

	assert.Equal(t, res.Code, http.StatusOK)
	assert.Equal(t,
		res.Body.String(),
		`{"return_value":1,"message":"Output of monitor_test_plugin. exit status is 1\n"}`,
	)
}

func TestMonitor3(t *testing.T) {
	// Critical

	m := martini.Classic()
	m.Use(render.Renderer())
	m.Post("/monitor", binding.Json(happo_agent.MonitorRequest{}), Monitor)

	reader := bytes.NewReader([]byte(`{
		"apikey": "",
		"plugin_name": "monitor_test_plugin",
		"plugin_option": "2"
	}`))
	req, _ := http.NewRequest("POST", "/monitor", reader)
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	lastRunned = time.Now().Unix() //avoid saveMachineState
	m.ServeHTTP(res, req)

	assert.Equal(t, res.Code, http.StatusOK)
	assert.Equal(t,
		res.Body.String(),
		`{"return_value":2,"message":"Output of monitor_test_plugin. exit status is 2\n"}`,
	)
}

func TestMonitor4(t *testing.T) {
	// Other

	m := martini.Classic()
	m.Use(render.Renderer())
	m.Post("/monitor", binding.Json(happo_agent.MonitorRequest{}), Monitor)

	reader := bytes.NewReader([]byte(`{
		"apikey": "",
		"plugin_name": "monitor_test_plugin",
		"plugin_option": "3"
	}`))
	req, _ := http.NewRequest("POST", "/monitor", reader)
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	lastRunned = time.Now().Unix() //avoid saveMachineState
	m.ServeHTTP(res, req)

	assert.Equal(t, res.Code, http.StatusOK)
	assert.Equal(t,
		res.Body.String(),
		`{"return_value":3,"message":"Output of monitor_test_plugin. exit status is 3\n"}`,
	)
}

func TestMonitor5(t *testing.T) {
	// Timeout

	m := martini.Classic()
	m.Use(render.Renderer())
	m.Post("/monitor", binding.Json(happo_agent.MonitorRequest{}), Monitor)

	reader := bytes.NewReader([]byte(fmt.Sprintf(`{
		"apikey": "",
		"plugin_name": "monitor_test_sleep",
		"plugin_option": "%d"
	}`, happo_agent.COMMAND_TIMEOUT+1)))
	req, _ := http.NewRequest("POST", "/monitor", reader)
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	lastRunned = time.Now().Unix() //avoid saveMachineState
	m.ServeHTTP(res, req)

	assert.Equal(t, res.Code, http.StatusServiceUnavailable)
	assert.Regexp(t,
		regexp.MustCompile(`^{"return_value":2,"message":"Exec timeout: .*monitor_test_sleep .*"}$`),
		res.Body.String(),
	)
}

func TestMonitor6(t *testing.T) {
	// plugin not found

	m := martini.Classic()
	m.Use(render.Renderer())
	m.Post("/monitor", binding.Json(happo_agent.MonitorRequest{}), Monitor)

	reader := bytes.NewReader([]byte(fmt.Sprintf(`{
		"apikey": "",
		"plugin_name": "notfound",
		"plugin_option": "%d"
	}`, happo_agent.COMMAND_TIMEOUT+1)))
	req, _ := http.NewRequest("POST", "/monitor", reader)
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	lastRunned = time.Now().Unix() //avoid saveMachineState
	m.ServeHTTP(res, req)

	assert.Equal(t, res.Code, http.StatusOK)
	assert.Equal(t,
		res.Body.String(),
		`{"return_value":127,"message":""}`,
	)
}
