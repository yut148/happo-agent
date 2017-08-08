package model

import (
	"net/http"
	"time"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/heartbeatsjp/happo-agent/collect"
	"github.com/heartbeatsjp/happo-agent/halib"
)

// --- Package Variables

// MetricConfigFile is filepath of metric config file
var MetricConfigFile string

// Metric returns collected metrics
func Metric(metricRequest halib.MetricRequest, r render.Render) {
	var metricResponse halib.MetricResponse

	metricResponse.MetricData = collect.GetCollectedMetricsWithLimit(60) // FIXME to prefer value. now 60 times = 1hour

	r.JSON(http.StatusOK, metricResponse)
}

// MetricAppend store metrics to local dbms
func MetricAppend(request halib.MetricAppendRequest, r render.Render) {
	var response halib.MetricAppendResponse

	err := collect.SaveMetrics(time.Now(), request.MetricData)
	if err != nil {
		response.Status = "error"
		response.Message = err.Error()
		r.JSON(http.StatusInternalServerError, response)
		return
	}

	response.Status = "ok"
	r.JSON(http.StatusOK, response)
}

// MetricConfigUpdate save metric collect config
func MetricConfigUpdate(metricRequest halib.MetricConfigUpdateRequest, r render.Render) {
	var metricResponse halib.MetricConfigUpdateResponse

	err := collect.SaveMetricConfig(metricRequest.Config, MetricConfigFile)
	if err != nil {
		metricResponse.Status = "NG"
	} else {
		metricResponse.Status = "OK"
	}

	r.JSON(http.StatusOK, metricResponse)
}

// MetricDataBufferStatus returns collected metrics status
func MetricDataBufferStatus(r render.Render) {
	r.JSON(http.StatusOK, collect.GetMetricDataBufferStatus())
}
