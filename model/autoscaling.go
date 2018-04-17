package model

import (
	"net/http"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/heartbeatsjp/happo-agent/autoscaling"
	"github.com/heartbeatsjp/happo-agent/halib"
)

// AutoScalingConfigFile is filepath of autoscaling config file
var AutoScalingConfigFile string

// AutoScalingConfigUpdate save autoscaling config
func AutoScalingConfigUpdate(autoScalingRequest halib.AutoScalingConfigUpdateRequest, r render.Render) {
	var autoScalingResponse halib.AutoScalingConfigUpdateResponse

	err := autoscaling.SaveAutoScalingConfig(autoScalingRequest.Config, AutoScalingConfigFile)
	if err != nil {
		autoScalingResponse.Status = "NG"
	} else {
		autoScalingResponse.Status = "OK"
	}

	r.JSON(http.StatusOK, autoScalingResponse)
}
