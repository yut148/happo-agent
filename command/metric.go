package command

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/codegangsta/cli"
	"github.com/heartbeatsjp/happo-agent/collect"
	"github.com/heartbeatsjp/happo-agent/lib"
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

	var metricsDataSlice []lib.MetricsData
	read, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	metricData, timestamp, err := collect.ParseMetricData(string(read))
	if err != nil {
		return err
	}

	var m lib.MetricsData
	m.HostName = hostname
	m.Timestamp = timestamp
	m.Metrics = metricData
	metricsDataSlice = append(metricsDataSlice, m)

	if dryRun {
		log.Println(metricsDataSlice)
		return nil
	}

	var metricAppendRequest lib.MetricAppendRequest

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
		log.Fatal(resp)
	}
	log.Printf("Success.")

	return nil
}
