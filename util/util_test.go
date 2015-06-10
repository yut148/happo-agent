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
