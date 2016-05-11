package util

import (
	"fmt"
	"testing"

	"github.com/heartbeatsjp/happo-lib"

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
}

func TestExecCommand2(t *testing.T) {
	command := "echo"
	option := "'hoge' >&2"

	exit_code, stdout, stderr, err := ExecCommand(command, option)
	assert.EqualValues(t, exit_code, 0)
	assert.Contains(t, stdout, "")
	assert.Contains(t, stderr, "hoge")
	assert.Nil(t, err)
}

func TestExecCommand3(t *testing.T) {
	command := "sleep"
	option := fmt.Sprintf("%d", happo_agent.COMMAND_TIMEOUT+1)

	exit_code, stdout, stderr, err := ExecCommand(command, option)
	assert.EqualValues(t, exit_code, -1)
	assert.Contains(t, stdout, "")
	assert.Contains(t, stderr, "")
	assert.NotNil(t, err)
}

func TestExecCommandCombinedOutput1(t *testing.T) {
	command := "echo"
	option := "'hoge'"

	exit_code, out, err := ExecCommandCombinedOutput(command, option)
	assert.EqualValues(t, exit_code, 0)
	assert.Contains(t, out, "hoge")
	assert.Nil(t, err)
}

func TestExecCommandCombinedOutput2(t *testing.T) {
	command := "echo"
	option := "'hoge' >&2"

	exit_code, out, err := ExecCommandCombinedOutput(command, option)
	assert.EqualValues(t, exit_code, 0)
	assert.Contains(t, out, "hoge")
	assert.Nil(t, err)
}

func TestExecCommandCombinedOutput3(t *testing.T) {
	command := "sleep"
	option := fmt.Sprintf("%d", happo_agent.COMMAND_TIMEOUT+1)

	exit_code, out, err := ExecCommandCombinedOutput(command, option)
	assert.EqualValues(t, exit_code, -1)
	assert.Contains(t, out, "")
	assert.NotNil(t, err)
}

func TestExecCommand4(t *testing.T) {
	command := "bash"
	option := "-c 'echo -n 1.STDOUT. ; echo -n 2.STDERR. >&2 ; echo -n 3.STDOUT. ; echo -n 4.STDERR. >&2 ; exit 0'"

	exit_code, out, err := ExecCommandCombinedOutput(command, option)
	assert.EqualValues(t, exit_code, 0)
	assert.Contains(t, out, "1.STDOUT.2.STDERR.3.STDOUT.4.STDERR.")
	assert.Nil(t, err)
}
