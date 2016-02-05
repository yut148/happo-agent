package model

import (
	"net/http"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/heartbeatsjp/happo-agent/collect"
	"github.com/heartbeatsjp/happo-lib"
)

// --- Package Variables
var MetricConfigFile string

func Metric(metric_request happo_agent.MetricRequest, r render.Render) {
	var metric_response happo_agent.MetricResponse

	metric_response.MetricData = collect.GetCollectedMetrics()

	r.JSON(http.StatusOK, metric_response)
}

func MetricConfigUpdate(metric_request happo_agent.MetricConfigUpdateRequest, r render.Render) {
	var metric_response happo_agent.MetricConfigUpdateResponse

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
