package model

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/heartbeatsjp/happo-agent/autoscaling"
	"github.com/heartbeatsjp/happo-agent/halib"
)

// AutoScalingConfigFile is filepath of autoscaling config file
var AutoScalingConfigFile string

// AutoScaling list autoscaling instances
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

	autoScalingList, err := autoscaling.GetAutoScalingConfig(AutoScalingConfigFile)
	if err != nil {
		response.Status = "error"
		response.Message = err.Error()
		r.JSON(http.StatusInternalServerError, response)
		return
	}

	var refreshAutoScalingGroups = []struct {
		autoScalingGroupName string
		autoScalingCount     int
		hostPrefix           string
	}{}
	for _, a := range autoScalingList.AutoScalings {
		if request.AutoScalingGroupName == a.AutoScalingGroupName || request.AutoScalingGroupName == "" {
			refreshAutoScalingGroups = append(refreshAutoScalingGroups, struct {
				autoScalingGroupName string
				autoScalingCount     int
				hostPrefix           string
			}{
				autoScalingGroupName: a.AutoScalingGroupName,
				autoScalingCount:     a.AutoScalingCount,
				hostPrefix:           a.HostPrefix,
			})
		}
		if request.AutoScalingGroupName == a.AutoScalingGroupName {
			break
		}
	}
	if len(refreshAutoScalingGroups) == 0 {
		response.Status = "error"
		response.Message = "can't find autoscaling group name in config"
		r.JSON(http.StatusNotFound, response)
		return
	}

	client := autoscaling.NewAWSClient()
	var errors []string
	for _, a := range refreshAutoScalingGroups {
		err = autoscaling.RefreshAutoScalingInstances(client, a.autoScalingGroupName, a.hostPrefix, a.autoScalingCount)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to refresh for %s: %s", a.autoScalingGroupName, err.Error()))
		}
	}
	if len(errors) > 0 {
		response.Status = "error"
		response.Message = strings.Join(errors, ",")
		r.JSON(http.StatusInternalServerError, response)
		return
	}

	response.Status = "OK"
	r.JSON(http.StatusOK, response)
}
