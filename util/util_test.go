package util

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/heartbeatsjp/happo-agent/lib"

	"github.com/stretchr/testify/assert"
)

func TestExecCommand1(t *testing.T) {
	command := "echo"
	option := "'hoge'"

	exit_code, stdout, stderr, err := ExecCommand(command, option)
	assert.EqualValues(t, exit_code, 0)
	assert.Contains(t, stdout, "hoge")
	assert.Contains(t, stderr, "")
	assert.Nil(t, err)

	_, ok := err.(*TimeoutError)
	assert.False(t, ok)
}

func TestExecCommand2(t *testing.T) {
	command := "echo"
	option := "'hoge' >&2"

	exit_code, stdout, stderr, err := ExecCommand(command, option)
	assert.EqualValues(t, exit_code, 0)
	assert.Contains(t, stdout, "")
	assert.Contains(t, stderr, "hoge")
	assert.Nil(t, err)

	_, ok := err.(*TimeoutError)
	assert.False(t, ok)
}

func TestExecCommand3(t *testing.T) {
	command := "sleep"
	option := fmt.Sprintf("%d", lib.COMMAND_TIMEOUT+1)

	exit_code, stdout, stderr, err := ExecCommand(command, option)
	assert.EqualValues(t, exit_code, -1)
	assert.Contains(t, stdout, "")
	assert.Contains(t, stderr, "")
	assert.NotNil(t, err)

	_, ok := err.(*TimeoutError)
	assert.True(t, ok)
}

func TestExecCommandCombinedOutput1(t *testing.T) {
	command := "echo"
	option := "'hoge'"

	exit_code, out, err := ExecCommandCombinedOutput(command, option)
	assert.EqualValues(t, exit_code, 0)
	assert.Contains(t, out, "hoge")
	assert.Nil(t, err)

	_, ok := err.(*TimeoutError)
	assert.False(t, ok)
}

func TestExecCommandCombinedOutput2(t *testing.T) {
	command := "echo"
	option := "'hoge' >&2"

	exit_code, out, err := ExecCommandCombinedOutput(command, option)
	assert.EqualValues(t, exit_code, 0)
	assert.Contains(t, out, "hoge")
	assert.Nil(t, err)

	_, ok := err.(*TimeoutError)
	assert.False(t, ok)
}

func TestExecCommandCombinedOutput3(t *testing.T) {
	command := "sleep"
	option := fmt.Sprintf("%d", lib.COMMAND_TIMEOUT+1)

	exit_code, out, err := ExecCommandCombinedOutput(command, option)
	assert.EqualValues(t, exit_code, -1)
	assert.Contains(t, out, "")
	assert.NotNil(t, err)

	_, ok := err.(*TimeoutError)
	assert.True(t, ok)
}

func TestExecCommand4(t *testing.T) {
	command := "bash"
	option := "-c 'echo -n 1.STDOUT. ; echo -n 2.STDERR. >&2 ; echo -n 3.STDOUT. ; echo -n 4.STDERR. >&2 ; exit 0'"

	exit_code, out, err := ExecCommandCombinedOutput(command, option)
	assert.EqualValues(t, exit_code, 0)
	assert.Contains(t, out, "1.STDOUT.2.STDERR.3.STDOUT.4.STDERR.")
	assert.Nil(t, err)

	_, ok := err.(*TimeoutError)
	assert.False(t, ok)
}

func TestBuildMetricAppendAPIRequest1(t *testing.T) {
	client, req, err := buildMetricAppendAPIRequest("https://127.0.0.2:6777", []byte(
		`{
		"api_key": "asdf",
		"metric_data":[
		{ "hostname":"saito-hb-vm101",
		"timestamp":1444028731,
		"metrics": {"linux.context_switches.context_switches":10001,
		"linux.disk.elapsed.iotime_sda":11,
		"linux.disk.elapsed.iotime_weighted_sda":111 }
	},
	{ "hostname":"saito-hb-vm102",
	"timestamp":1444028732,
	"metrics":{
		"linux.context_switches.context_switches":20002,
		"linux.disk.elapsed.iotime_sda":22,
		"linux.disk.elapsed.iotime_weighted_sda":222 }
	}
	]}`))
	assert.True(t, (client.Transport.(*http.Transport)).TLSClientConfig.InsecureSkipVerify)
	assert.Equal(t, req.URL.Scheme, "https")
	assert.Equal(t, req.URL.Host, "127.0.0.2:6777")
	assert.Equal(t, req.URL.Path, "/metric/append")
	assert.Equal(t, req.Method, "POST")
	assert.Equal(t, req.Header.Get("Content-Type"), "application/json")
	assert.Nil(t, err)
}
