package util

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
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

	return exitStatus.GetChildExitCode(), stdout, stderr, err
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
