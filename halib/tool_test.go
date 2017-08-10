package halib

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const HOST = "10.0.0.1"
const PORT = DefaultAgentPort
const METHOD = "TEST"
const JSON = "{\"test\":100}"

func TestGetProxyJSON1(t *testing.T) {
	ProxyHosts := []string{"192.168.0.1"}
	var jsonData ProxyRequest
	ProxyRequest := ProxyRequest{
		ProxyHostPort: []string{fmt.Sprintf("%s:%d", HOST, PORT)},
		RequestType:   METHOD,
		RequestJSON:   ([]byte)(JSON),
	}

	jsonStr, agentHost, agentPort, err := GetProxyJSON(ProxyHosts, HOST, PORT, "TEST", ([]byte(JSON)))
	assert.Nil(t, err)
	assert.EqualValues(t, "192.168.0.1", agentHost)
	assert.EqualValues(t, DefaultAgentPort, agentPort)

	json.Unmarshal(jsonStr, &jsonData)
	assert.EqualValues(t, ProxyRequest, jsonData)
}

func TestGetProxyJSON2(t *testing.T) {
	ProxyHosts := []string{"192.168.0.1", "172.16.0.1"}
	var jsonData ProxyRequest
	ProxyRequest := ProxyRequest{
		ProxyHostPort: []string{"172.16.0.1", fmt.Sprintf("%s:%d", HOST, PORT)},
		RequestType:   METHOD,
		RequestJSON:   ([]byte)(JSON),
	}

	jsonStr, agentHost, agentPort, err := GetProxyJSON(ProxyHosts, HOST, PORT, "TEST", ([]byte(JSON)))
	assert.Nil(t, err)
	assert.EqualValues(t, "192.168.0.1", agentHost)
	assert.EqualValues(t, DefaultAgentPort, agentPort)

	json.Unmarshal(jsonStr, &jsonData)
	assert.EqualValues(t, ProxyRequest, jsonData)
}
