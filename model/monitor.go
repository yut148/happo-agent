package model

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/heartbeatsjp/happo-agent/db"
	"github.com/heartbeatsjp/happo-agent/halib"
	"github.com/heartbeatsjp/happo-agent/util"
)

// --- Constant Values

const errorLogCommands = "w,ps auxwwf,ss -anp,lsof"
const errorLogOutputPath = "/tmp"
const errorLogOutputFilename = "snapshot_%s.log"

var (
	saveStateChan   = make(chan bool)
	lastRunnedMutex = sync.Mutex{}
	lastRunned      int64
	// ErrorLogIntervalSeconds is error log collect interval
	ErrorLogIntervalSeconds = int64(halib.DefaultErrorLogIntervalSeconds)
	// NagiosPluginPaths is nagios plugin search paths. combined with `,`
	NagiosPluginPaths = halib.DefaultNagiosPluginPaths
)

// --- Method

func init() {
	lastRunned = 0
	go func() {
		for {
			select {
			case <-saveStateChan:
				go func() {
					if isPermitSaveState() {
						err := saveMachineState()
						if err != nil {
							log := util.HappoAgentLogger()
							log.Errorf("while saveMachieState(): %s, %v", err, time.Now())
						}
					}
				}()
			}
		}
	}()
}

// Monitor execute monitor command and returns result
func Monitor(monitorRequest halib.MonitorRequest, r render.Render) {
	log := util.HappoAgentLogger()
	var monitorResponse halib.MonitorResponse

	if !util.Production {
		log.Println(fmt.Sprintf("Plugin Name: %s, Option: %s", monitorRequest.PluginName, monitorRequest.PluginOption))
	}
	ret, message, err := execPluginCommand(monitorRequest.PluginName, monitorRequest.PluginOption)
	if err != nil {
		monitorResponse.ReturnValue = halib.MonitorError
		monitorResponse.Message = err.Error()
		//if _, ok := err.(*util.TimeoutError); ok {
		//	r.JSON(http.StatusInternalServerError, monitorResponse)
		//	return
		//}
		r.JSON(http.StatusInternalServerError, monitorResponse)
		return
	}
	if ret != 0 {
		saveStateChan <- true
	}

	monitorResponse.ReturnValue = ret
	monitorResponse.Message = message

	r.JSON(http.StatusOK, monitorResponse)
}

func execPluginCommand(pluginName string, pluginOption string) (int, string, error) {
	log := util.HappoAgentLogger()
	var plugin string

	for _, basePath := range strings.Split(NagiosPluginPaths, ",") {
		plugin = path.Join(basePath, pluginName)
		_, err := os.Stat(plugin)
		if err == nil {
			if !util.Production {
				log.Println(plugin)
			}
			break
		}
	}

	exitstatus, stdout, stderr, err := util.ExecCommand(plugin, pluginOption)

	out := stdout
	if stdout == "" {
		out = fmt.Sprintf(`stdout=, stderr=%s`, stderr)
	} else if stderr != "" {
		out = fmt.Sprintf("%s, stderr=%s", stdout, stderr)
	}

	return exitstatus, out, err
}

func saveMachineState() error {
	var err error
	log := util.HappoAgentLogger()
	loggedTime := time.Now()

	result := ""
	for _, cmd := range strings.Split(errorLogCommands, ",") {
		cmd := strings.Split(cmd, " ")
		if len(cmd) == 1 {
			cmd = append(cmd, "")
		}
		exitstatus, stdout, _, err := util.ExecCommand(cmd[0], cmd[1])
		if exitstatus == 0 && err == nil {
			result += fmt.Sprintf("********** %s %s (%s) **********\n", cmd[0], cmd[1], loggedTime.Format(time.RFC3339))
			result += stdout
			result += "\n\n"
		}
	}

	err = db.DB.Update(func(tx *bolt.Tx) error {
		bucket := db.MachineStateBucket(tx)
		err := bucket.Put(
			db.TimeToKey(loggedTime),
			[]byte(result))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = db.DB.Update(func(tx *bolt.Tx) error {
		// retire old metrics
		bucket := db.MachineStateBucket(tx)
		cursor := bucket.Cursor()

		oldestThreshold := loggedTime.Add(time.Duration(-1*db.MachineStateMaxLifetimeSeconds) * time.Second)

		retrivedKeys := make([][]byte, 0)
		oldestThresholdKey := db.TimeToKey(oldestThreshold)
		for key, _ := cursor.First(); key != nil && bytes.Compare(key, oldestThresholdKey) <= 0; key, _ = cursor.Next() {
			retrivedKeys = append(retrivedKeys, key)

			// logging
			log.Warn("retire old metrics: key=%v\n", string(key))
			// if write value to log, log become too large...
		}

		for _, key := range retrivedKeys {
			bucket.Delete(key)
		}
		return nil

	})
	if err != nil {
		return err
	}

	return nil
}

func isPermitSaveState() bool {
	log := util.HappoAgentLogger()
	if ErrorLogIntervalSeconds < 0 {
		return false
	}

	lastRunnedMutex.Lock()
	defer lastRunnedMutex.Unlock()

	duration := time.Now().Unix() - lastRunned
	if duration < ErrorLogIntervalSeconds {
		log.Debug(fmt.Sprintf("Duration: %d < %d", duration, ErrorLogIntervalSeconds))
		return false
	}
	lastRunned = time.Now().Unix()
	return true
}
