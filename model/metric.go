package model

import (
	"net/http"
	"time"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/heartbeatsjp/happo-agent/collect"
	"github.com/heartbeatsjp/happo-agent/lib"
)

// --- Package Variables
var MetricConfigFile string

func Metric(metric_request lib.MetricRequest, r render.Render) {
	var metric_response lib.MetricResponse

	metric_response.MetricData = collect.GetCollectedMetricsWithLimit(60) // FIXME to prefer value. now 60 times = 1hour

	r.JSON(http.StatusOK, metric_response)
}

//MetricAppend store metrics to local dbms
func MetricAppend(request lib.MetricAppendRequest, r render.Render) {
	var response lib.MetricAppendResponse

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

func MetricConfigUpdate(metric_request lib.MetricConfigUpdateRequest, r render.Render) {
	var metric_response lib.MetricConfigUpdateResponse

	err := collect.SaveMetricConfig(metric_request.Config, MetricConfigFile)
	if err != nil {
		metric_response.Status = "NG"
	} else {
		metric_response.Status = "OK"
	}

	r.JSON(http.StatusOK, metric_response)
}

func MetricDataBufferStatus(r render.Render) {
	r.JSON(http.StatusOK, collect.GetMetricDataBufferStatus())
}
