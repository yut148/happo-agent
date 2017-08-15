package model

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/heartbeatsjp/happo-agent/db"
	"github.com/heartbeatsjp/happo-agent/halib"
	"github.com/heartbeatsjp/happo-agent/util"
	leveldbUtil "github.com/syndtr/goleveldb/leveldb/util"
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
	ErrorLogIntervalSeconds int64
)

// --- Method

func init() {
	if ErrorLogIntervalSeconds == 0 {
		//unconfigured
		ErrorLogIntervalSeconds = halib.DefaultErrorLogIntervalSeconds
	}
	lastRunned = 0
	go func() {
		for {
			select {
			case <-saveStateChan:
				go func() {
					if isPermitSaveState() {
						err := saveMachineState()
						if err != nil {
							log.Println(fmt.Sprintf("ERROR while saveMachieState(): %s, %v", err, time.Now()))
						}
					}
				}()
			}
		}
	}()
}

// Monitor execute monitor command and returns result
func Monitor(monitorRequest halib.MonitorRequest, r render.Render) {
	var monitorResponse halib.MonitorResponse

	if !util.Production {
		log.Println(fmt.Sprintf("Plugin Name: %s, Option: %s", monitorRequest.PluginName, monitorRequest.PluginOption))
	}
	ret, message, err := execPluginCommand(monitorRequest.PluginName, monitorRequest.PluginOption)
	if err != nil {
		monitorResponse.ReturnValue = halib.MonitorError
		monitorResponse.Message = err.Error()
		if _, ok := err.(*util.TimeoutError); ok {
			r.JSON(http.StatusServiceUnavailable, monitorResponse)
			return
		}
		r.JSON(http.StatusBadRequest, monitorResponse)
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
	var plugin string

	for _, basePath := range strings.Split(halib.DefaultNagiosPluginPaths, ",") {
		plugin = path.Join(basePath, pluginName)
		_, err := os.Stat(plugin)
		if err == nil {
			if !util.Production {
				log.Println(plugin)
			}
			break
		}
	}

	exitstatus, stdout, _, err := util.ExecCommand(plugin, pluginOption)

	if err != nil {
		return halib.MonitorUnknown, "", err
	}

	return exitstatus, stdout, nil
}

func saveMachineState() error {
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

	transaction, err := db.DB.OpenTransaction()
	if err != nil {
		return err
	}

	transaction.Put(
		[]byte(fmt.Sprintf("s-%d", loggedTime.Unix())),
		[]byte(result),
		nil)
	err = transaction.Commit()

	if err != nil {
		return err
	}

	// retire old metrics
	transaction, err = db.DB.OpenTransaction()
	if err != nil {
		log.Println(err)
	}
	oldestThreshold := loggedTime.Add(time.Duration(-1*db.MachineStateMaxLifetimeSeconds) * time.Second)
	iter := transaction.NewIterator(
		&leveldbUtil.Range{
			Start: []byte("s-0"),
			Limit: []byte(fmt.Sprintf("s-%d", oldestThreshold.Unix()))},
		nil)
	for iter.Next() {
		key := iter.Key()
		transaction.Delete(key, nil)

		// logging
		unixTime, _ := strconv.Atoi(strings.SplitN(string(key), "-", 2)[1])
		log.Printf("retire old metrics: key=%v(%v)\n", string(key), time.Unix(int64(unixTime), 0))
		// if write value to log, log become too large...
	}
	iter.Release()

	err = transaction.Commit()
	if err != nil {
		return err
	}

	return nil
}

func isPermitSaveState() bool {
	if ErrorLogIntervalSeconds < 0 {
		return false
	}

	lastRunnedMutex.Lock()
	defer lastRunnedMutex.Unlock()

	duration := time.Now().Unix() - lastRunned
	if duration < ErrorLogIntervalSeconds {
		log.Println(fmt.Sprintf("Duration: %d < %d", duration, ErrorLogIntervalSeconds))
		return false
	}
	lastRunned = time.Now().Unix()
	return true
}
