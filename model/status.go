package model

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/heartbeatsjp/happo-agent/collect"
	"github.com/heartbeatsjp/happo-agent/halib"
	"github.com/heartbeatsjp/happo-agent/util"
)

var (
	// AppVersion equals main.Version
	AppVersion string

	startAt = time.Now()
)

// Status implements /status endpoint. returns status
func Status(req *http.Request, r render.Render) {
	log := util.HappoAgentLogger()

	callers := make([]string, 0)
	pcs := make([]uintptr, runtime.NumGoroutine())
	runtime.Callers(0, pcs)
	for _, pc := range pcs {
		f := runtime.FuncForPC(pc)
		filepath, line := f.FileLine(pc)
		log.Debug(fmt.Sprintf("%s:%d", filepath, line))
		callers = append(callers, fmt.Sprintf("%s:%d", filepath, line))
	}

	statusResponse := &halib.StatusResponse{
		AppVersion:         AppVersion,
		UptimeSeconds:      int64(time.Since(startAt) / time.Second),
		NumGoroutine:       runtime.NumGoroutine(),
		MetricBufferStatus: collect.GetMetricDataBufferStatus(false),
		Callers:            callers,
	}
	r.JSON(http.StatusOK, statusResponse)
}

// RequestStatus implements /status/request endpoint. returns status
func RequestStatus(req *http.Request, r render.Render) {
	requestStatus := util.GetMartiniRequestStatus(time.Now())

	r.JSON(http.StatusOK, requestStatus)
}

// MemoryStatus implements /status/memory endpoint. returns runtime.Memstatus in JSON
func MemoryStatus(req *http.Request, r render.Render) {
	mem := new(runtime.MemStats)
	runtime.ReadMemStats(mem)
	r.JSON(http.StatusOK, mem)
}
