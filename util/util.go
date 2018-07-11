package util

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/heartbeatsjp/happo-agent/halib"

	"github.com/Songmu/timeout"
	"github.com/codegangsta/cli"
)

// Global Variables

// CommandTimeout is command execution timeout sec
var CommandTimeout time.Duration = -1

// Production is flag. when production use, set true
var Production bool

// TimeoutError is error struct show error is timeout
type TimeoutError struct {
	Message string
}

func (err *TimeoutError) Error() string {
	return err.Message
}

// --- Function
func init() {
	Production = strings.ToLower(os.Getenv("MARTINI_ENV")) == "production"
}

// ExecCommand execute command with specified timeout behavior
func ExecCommand(command string, option string) (int, string, string, error) {
	var timeBegin time.Time
	var cswBegin int
	if HappoAgentLoggerEnableInfo() {
		timeBegin = time.Now()
		cswBegin = getContextSwitch()
	}

	commandTimeout := CommandTimeout
	if commandTimeout == -1 {
		commandTimeout = halib.DefaultCommandTimeout
	}

	commandWithOptions := fmt.Sprintf("%s %s", command, option)
	tio := &timeout.Timeout{
		Cmd:       exec.Command("/bin/sh", "-c", commandWithOptions),
		Duration:  commandTimeout * time.Second,
		KillAfter: halib.CommandKillAfterSeconds * time.Second,
	}
	exitStatus, stdout, stderr, err := tio.Run()

	if err == nil && exitStatus.IsTimedOut() {
		err = &TimeoutError{"Exec timeout: " + commandWithOptions}
	}

	if HappoAgentLoggerEnableInfo() {
		now := time.Now()
		cswTook := getContextSwitch() - cswBegin
		timeTook := now.Sub(timeBegin)
		HappoAgentLogger().Infof("%v: ExecCommand %v end. csw=%v, duration=%v,", now.Format(time.RFC3339Nano), command, cswTook, timeTook.Seconds())
	}
	return exitStatus.GetChildExitCode(), stdout, stderr, err
}

func getContextSwitch() int {
	if _, err := os.Stat("/proc/stat"); err != nil {
		return -1
	}
	fp, err := os.Open("/proc/stat")
	defer fp.Close()
	if err != nil {
		return -1
	}
	for scanner := bufio.NewScanner(fp); scanner.Scan(); {
		line := scanner.Text()
		if strings.HasPrefix(line, "ctxt ") {
			csw, err := strconv.Atoi(strings.Split(line, " ")[1])
			if err != nil {
				return -1
			}
			return csw
		}
	}
	return -1
}

// ExecCommandCombinedOutput execute command with specified timeout behavior
func ExecCommandCombinedOutput(command string, option string) (int, string, error) {

	commandTimeout := CommandTimeout
	if commandTimeout == -1 {
		commandTimeout = halib.DefaultCommandTimeout
	}

	commandWithOptions := fmt.Sprintf("%s %s", command, option)
	tio := &timeout.Timeout{
		Cmd:       exec.Command("/bin/sh", "-c", commandWithOptions),
		Duration:  commandTimeout * time.Second,
		KillAfter: halib.CommandKillAfterSeconds * time.Second,
	}
	out := &bytes.Buffer{}
	tio.Cmd.Stdout = out
	tio.Cmd.Stderr = out

	ch, err := tio.RunCommand()
	exitStatus := <-ch

	if err == nil && exitStatus.IsTimedOut() {
		err = &TimeoutError{"Exec timeout: " + commandWithOptions}
	}

	return exitStatus.GetChildExitCode(), out.String(), err

}

// BindManageParameter build and return ManageRequest
func BindManageParameter(c *cli.Context) (halib.ManageRequest, error) {
	var hostinfo halib.CrawlConfigAgent
	var manageRequest halib.ManageRequest

	hostinfo.GroupName = c.String("group_name")
	if hostinfo.GroupName == "" {
		return manageRequest, errors.New("group_name is null")
	}
	hostinfo.IP = c.String("ip")
	if hostinfo.GroupName == "" {
		return manageRequest, errors.New("ip is null")
	}
	hostinfo.Hostname = c.String("hostname")
	hostinfo.Port = c.Int("port")
	hostinfo.Proxies = c.StringSlice("proxy")
	manageRequest.Hostdata = hostinfo

	return manageRequest, nil
}

// RequestToManageAPI send request to ManageAPI
func RequestToManageAPI(endpoint string, path string, postdata []byte) (*http.Response, error) {
	uri := fmt.Sprintf("%s%s", endpoint, path)
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(postdata))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return http.DefaultTransport.RoundTrip(req)
}

// RequestToMetricAppendAPI send request to MetricAppendPI
func RequestToMetricAppendAPI(endpoint string, postdata []byte) (*http.Response, error) {
	client, req, err := buildMetricAppendAPIRequest(endpoint, postdata)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

func buildMetricAppendAPIRequest(endpoint string, postdata []byte) (*http.Client, *http.Request, error) {
	uri := fmt.Sprintf("%s/metric/append", endpoint)
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(postdata))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	//FIXME other parameters should be proper values
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	return client, req, err
}

// RequestToAutoScalingResolveAPI send request to AutoScalingResolveAPI
func RequestToAutoScalingResolveAPI(endpoint string, alias string) (*http.Response, error) {
	client, req, err := buildAutoScalingResolveAPIRequest(endpoint, alias)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

func buildAutoScalingResolveAPIRequest(endpoint string, alias string) (*http.Client, *http.Request, error) {
	uri := fmt.Sprintf("%s/autoscaling/resolve/%s", endpoint, alias)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, nil, err
	}

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	return client, req, err
}
