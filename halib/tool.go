package halib

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// GetProxyJSON returns Marshal-ed ProxyRequest in []byte
func GetProxyJSON(proxyHosts []string, host string, port int, requestType string, proxyJSONStr []byte) ([]byte, string, int, error) {
	var agentHost string
	var agentPort int
	var err error

	// Step 1
	agentHostport := strings.Split(proxyHosts[0], ":")
	agentHost = agentHostport[0]
	if len(agentHostport) == 2 {
		agentPort, err = strconv.Atoi(agentHostport[1])
		if err != nil {
			return nil, "", 0, err
		}
	} else {
		agentPort = DefaultAgentPort
	}

	// Step 2 or later
	proxyHosts = proxyHosts[1:]
	proxyHosts = append(proxyHosts, fmt.Sprintf("%s:%d", host, port))

	proxyRequest := ProxyRequest{
		ProxyHostPort: proxyHosts,
		RequestType:   requestType,
		RequestJSON:   proxyJSONStr,
	}
	jsonData, _ := json.Marshal(proxyRequest)

	return jsonData, agentHost, agentPort, nil
}
