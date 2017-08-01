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

	"github.com/heartbeatsjp/happo-agent/lib"

	"github.com/Songmu/timeout"
	"github.com/codegangsta/cli"
)

// Global Variables
var CommandTimeout time.Duration = -1
var Production bool

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

func ExecCommand(command string, option string) (int, string, string, error) {

	command_timeout := CommandTimeout
	if command_timeout == -1 {
		command_timeout = lib.COMMAND_TIMEOUT
	}

	command_with_options := fmt.Sprintf("%s %s", command, option)
	tio := &timeout.Timeout{
		Cmd:       exec.Command("/bin/sh", "-c", command_with_options),
		Duration:  command_timeout * time.Second,
		KillAfter: lib.COMMAND_KILLAFTER * time.Second,
	}
	exitStatus, stdout, stderr, err := tio.Run()

	if err == nil && exitStatus.IsTimedOut() {
		err = &TimeoutError{"Exec timeout: " + command_with_options}
	}

	return exitStatus.GetChildExitCode(), stdout, stderr, err
}

func ExecCommandCombinedOutput(command string, option string) (int, string, error) {

	command_timeout := CommandTimeout
	if command_timeout == -1 {
		command_timeout = lib.COMMAND_TIMEOUT
	}

	command_with_options := fmt.Sprintf("%s %s", command, option)
	tio := &timeout.Timeout{
		Cmd:       exec.Command("/bin/sh", "-c", command_with_options),
		Duration:  command_timeout * time.Second,
		KillAfter: lib.COMMAND_KILLAFTER * time.Second,
	}
	out := &bytes.Buffer{}
	tio.Cmd.Stdout = out
	tio.Cmd.Stderr = out

	ch, err := tio.RunCommand()
	exitStatus := <-ch

	if err == nil && exitStatus.IsTimedOut() {
		err = &TimeoutError{"Exec timeout: " + command_with_options}
	}

	return exitStatus.GetChildExitCode(), out.String(), err

}

func BindManageParameter(c *cli.Context) (lib.ManageRequest, error) {
	var hostinfo lib.CrawlConfigAgent
	var manage_request lib.ManageRequest

	hostinfo.GroupName = c.String("group_name")
	if hostinfo.GroupName == "" {
		return manage_request, errors.New("group_name is null")
	}
	hostinfo.IP = c.String("ip")
	if hostinfo.GroupName == "" {
		return manage_request, errors.New("ip is null")
	}
	hostinfo.Hostname = c.String("hostname")
	hostinfo.Port = c.Int("port")
	hostinfo.Proxies = c.StringSlice("proxy")
	manage_request.Hostdata = hostinfo

	return manage_request, nil
}

func RequestToManageAPI(endpoint string, path string, postdata []byte) (*http.Response, error) {
	uri := fmt.Sprintf("%s%s", endpoint, path)
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(postdata))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return http.DefaultTransport.RoundTrip(req)
}

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
