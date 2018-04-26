package model

import (
	"net/http"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/heartbeatsjp/happo-agent/autoscaling"
	"github.com/heartbeatsjp/happo-agent/halib"
)

// AutoScalingConfigFile is filepath of autoscaling config file
var AutoScalingConfigFile string

func AutoScaling(req *http.Request, r render.Render) {
	var autoScalingResponse halib.AutoScalingResponse

	autoScaling, err := autoscaling.AutoScaling(AutoScalingConfigFile)
	if err != nil {
		r.JSON(http.StatusInternalServerError, autoScalingResponse)
		return
	}
	autoScalingResponse.AutoScaling = autoScaling
	r.JSON(http.StatusOK, autoScalingResponse)
}

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

// AutoScalingRefresh refresh autoscaling
func AutoScalingRefresh(request halib.AutoScalingRefreshRequest, r render.Render) {
	var response halib.AutoScalingRefreshResponse

	if request.AutoScalingGroupName == "" {
		response.Status = "error"
		response.Message = "autoscaling_group_name required"
		r.JSON(http.StatusBadRequest, response)
		return
	}

	autoScalingList, err := autoscaling.GetAutoScalingConfig(AutoScalingConfigFile)
	if err != nil {
		response.Status = "error"
		response.Message = err.Error()
		r.JSON(http.StatusInternalServerError, response)
		return
	}

	var (
		autoScalingGroupName string
		autoScalingCount     int
		hostPrefix           string
	)
	for _, a := range autoScalingList.AutoScalings {
		if request.AutoScalingGroupName == a.AutoScalingGroupName {
			autoScalingGroupName = a.AutoScalingGroupName
			autoScalingCount = a.AutoScalingCount
			hostPrefix = a.HostPrefix
			break
		}
	}

	client := autoscaling.NewAWSClient()
	err = autoscaling.RefreshAutoScalingInstances(client, autoScalingGroupName, hostPrefix, autoScalingCount)
	if err != nil {
		response.Status = "error"
		response.Message = err.Error()
		r.JSON(http.StatusOK, response)
		return
	}

	response.Status = "OK"
	r.JSON(http.StatusOK, response)
}
