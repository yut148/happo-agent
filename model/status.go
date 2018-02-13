package model

import (
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/heartbeatsjp/happo-agent/collect"
	"github.com/heartbeatsjp/happo-agent/db"
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

	bufSize := 4 * 1024 * 1024 // max 4MB
	buf := make([]byte, bufSize)
	bufferOverflow := false
	readBytes := runtime.Stack(buf, true)
	if readBytes < len(buf) {
		buf = buf[:readBytes] // shrink
	} else {
		// note: strictly saying, in case stack is just same as bufSize, buffer is not overflow.
		bufferOverflow = true
	}
	callers := strings.Split(string(buf), "\n\n")
	if bufferOverflow {
		callers = append(callers, "...")
	}
	log.Debugf("callers: %v", callers)

	boltDBStats := map[string]int{}
	boltDBStats["FreePageN"] = db.DB.Stats().FreePageN
	boltDBStats["PendingPageN"] = db.DB.Stats().PendingPageN
	boltDBStats["FreeAlloc"] = db.DB.Stats().FreeAlloc
	boltDBStats["FreelistInuse"] = db.DB.Stats().FreelistInuse
	boltDBStats["TxN"] = db.DB.Stats().TxN
	boltDBStats["OpenTxN"] = db.DB.Stats().OpenTxN
	log.Debugf("boltDBStats: %v", boltDBStats)

	statusResponse := &halib.StatusResponse{
		AppVersion:         AppVersion,
		UptimeSeconds:      int64(time.Since(startAt) / time.Second),
		NumGoroutine:       runtime.NumGoroutine(),
		MetricBufferStatus: collect.GetMetricDataBufferStatus(false),
		Callers:            callers,
		LevelDBProperties:  map[string]string{},
		BoltDBStats:        boltDBStats,
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
