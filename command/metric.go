package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/codegangsta/cli"
	"github.com/heartbeatsjp/happo-agent/collect"
	"github.com/heartbeatsjp/happo-agent/halib"
	"github.com/heartbeatsjp/happo-agent/util"
)

//CmdAppendMetric is action of subcommand append_metric
func CmdAppendMetric(c *cli.Context) error {
	var err error

	hostname := c.String("hostname")
	bastionEndoint := c.String("bastion-endpoint")
	datafileArg := c.String("datafile")
	dryRun := c.Bool("dry-run")

	var f *os.File
	defer f.Close()
	if datafileArg == "-" {
		f = os.Stdin
	} else {
		f, err = os.Open(datafileArg)
		if err != nil {
			return err
		}
	}

	var metricsDataSlice []halib.MetricsData
	read, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	metricData, timestamp, err := collect.ParseMetricData(string(read))
	if err != nil {
		return err
	}

	var m halib.MetricsData
	m.HostName = hostname
	m.Timestamp = timestamp
	m.Metrics = metricData
	metricsDataSlice = append(metricsDataSlice, m)

	if dryRun {
		fmt.Println(metricsDataSlice)
		return nil
	}

	var metricAppendRequest halib.MetricAppendRequest

	metricAppendRequest.APIKey = c.String("api-key")
	metricAppendRequest.MetricData = metricsDataSlice

	data, err := json.Marshal(metricAppendRequest)
	if err != nil {
		return err
	}

	resp, err := util.RequestToMetricAppendAPI(bastionEndoint, data)
	if err != nil && resp == nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	fmt.Println("Success.")

	return nil
}
