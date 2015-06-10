package model

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/heartbeatsjp/happo-agent/util"
	"github.com/heartbeatsjp/happo-lib"
)

// --- Constant Values

const ERROR_LOG_COMMANDS = "w,ps auxwwf,ss -anp,lsof"
const ERROR_LOG_OUTPUT_PATH = "/tmp"
const ERROR_LOG_OUTPUT_FILENAME = "snapshot_%s.log"

// --- Method

func Monitor(monitor_request happo_agent.MonitorRequest, r render.Render) {
	var monitor_response happo_agent.MonitorResponse

	log.Println(fmt.Sprintf("Plugin Name: %s, Option: %s", monitor_request.Plugin_Name, monitor_request.Plugin_Option))
	ret, message, err := execPluginCommand(monitor_request.Plugin_Name, monitor_request.Plugin_Option)
	if err != nil {
		monitor_response.Return_Value = happo_agent.MONITOR_UNKNOWN
		r.JSON(http.StatusBadRequest, monitor_response)
		return
	}
	if ret != 0 {
		saveMachineState()
	}

	monitor_response.Return_Value = ret
	monitor_response.Message = message

	r.JSON(http.StatusOK, monitor_response)
}

func execPluginCommand(plugin_name string, plugin_option string) (int, string, error) {
	var plugin string

	for _, base_path := range strings.Split(happo_agent.NAGIOS_PLUGIN_PATHS, ",") {
		plugin = path.Join(base_path, plugin_name)
		_, err := os.Stat(plugin)
		if err == nil {
			log.Println(plugin)
			break
		}
	}

	exitstatus, stdout, _, err := util.ExecCommand(plugin, plugin_option)

	if err != nil {
		return happo_agent.MONITOR_UNKNOWN, "", err
	}

	return exitstatus, stdout, nil
}

func saveMachineState() error {
	if saveState() {
		return nil
	}
	logged_time := time.Now()

	result := ""
	for _, cmd := range strings.Split(ERROR_LOG_COMMANDS, ",") {
		cmd := strings.Split(cmd, " ")
		if len(cmd) == 1 {
			cmd = append(cmd, "")
		}
		exitstatus, stdout, _, err := util.ExecCommand(cmd[0], cmd[1])
		if exitstatus == 0 && err == nil {
			result += fmt.Sprintf("********** %s %s (%s) **********\n", cmd[0], cmd[1], logged_time.Format(time.RFC3339))
			result += stdout
			result += "\n\n"
		}
	}

	file_suffix_time := fmt.Sprintf("%04d%02d%02d_%02d%02d%02d", logged_time.Year(), logged_time.Month(), logged_time.Day(), logged_time.Hour(), logged_time.Minute(), logged_time.Second())
	filepath := path.Join(ERROR_LOG_OUTPUT_PATH, fmt.Sprintf(ERROR_LOG_OUTPUT_FILENAME, file_suffix_time))
	err := ioutil.WriteFile(filepath, []byte(result), os.ModePerm)

	return err
}

func isPermitSaveState() func() bool {
	var last_runned int64
	last_runned = 0
	return func() bool {
		duration := time.Now().Unix() - last_runned
		if duration < happo_agent.ERROR_LOG_INTERVAL_SEC {
			log.Println(fmt.Sprintf("Duration: %d < %d", duration, happo_agent.ERROR_LOG_INTERVAL_SEC))
			return true
		}
		last_runned = time.Now().Unix()
		return false
	}
}

var saveState = isPermitSaveState()
