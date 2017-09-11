package model

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/heartbeatsjp/happo-agent/collect"
	"github.com/heartbeatsjp/happo-agent/util"
)

var (
	// AppVersion equals main.Version
	AppVersion string

	startAt = time.Now()
)

// AgentStatus is happo-agent status
type AgentStatus struct {
	AppVersion         string
	UptimeSeconds      int64
	NumGoroutine       int
	MetricBufferStatus map[string]int64
	Callers            []string
}

// ExtendedAgentStatus is extended-happo-agent status
type ExtendedAgentStatus struct {
	AgentStatus
	MemStatus *runtime.MemStats
}

// Status implements /status endpoint. returns status
func Status(req *http.Request, r render.Render) {
	log := util.HappoAgentLogger()

	extended := false
	for key := range req.URL.Query() {
		if strings.ToLower(key) == "extended" {
			extended = true
			break
		}
	}

	callers := make([]string, 0)
	pcs := make([]uintptr, runtime.NumGoroutine())
	runtime.Callers(0, pcs)
	for _, pc := range pcs {
		f := runtime.FuncForPC(pc)
		filepath, line := f.FileLine(pc)
		log.Debug(fmt.Sprintf("%s:%d", filepath, line))
		callers = append(callers, fmt.Sprintf("%s:%d", filepath, line))
	}

	agentStatus := &AgentStatus{
		AppVersion:         AppVersion,
		UptimeSeconds:      int64(time.Since(startAt) / time.Second),
		NumGoroutine:       runtime.NumGoroutine(),
		MetricBufferStatus: collect.GetMetricDataBufferStatus(extended),
		Callers:            callers,
	}
	if extended {
		mem := new(runtime.MemStats)
		runtime.ReadMemStats(mem)

		extendedAgentStatus := &ExtendedAgentStatus{
			*agentStatus,
			mem,
		}
		r.JSON(http.StatusOK, extendedAgentStatus)
	} else {
		r.JSON(http.StatusOK, agentStatus)
	}
}

// RequestStatus implements /status/request endpoint. returns status
func RequestStatus(req *http.Request, r render.Render) {
	requestStatus := util.GetMartiniRequestStatus(time.Now())

	r.JSON(http.StatusOK, requestStatus)
}
