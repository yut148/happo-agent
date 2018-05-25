package collect

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/heartbeatsjp/happo-agent/db"
	"github.com/heartbeatsjp/happo-agent/halib"
	"github.com/heartbeatsjp/happo-agent/util"
	leveldbErrors "github.com/syndtr/goleveldb/leveldb/errors"
	leveldbUtil "github.com/syndtr/goleveldb/leveldb/util"

	"gopkg.in/yaml.v2"
)

var (
	// SensuPluginPaths is sensu plugin search paths. combined with `,`
	SensuPluginPaths = halib.DefaultSensuPluginPaths
)

// --- Method

// Metrics is main function of metric collection
func Metrics(configPath string) error {
	var metricsDataBuffer []halib.MetricsData

	metricList, err := GetMetricConfig(configPath)
	if err != nil {
		return err
	}

	metricTotalCount := 0
	for _, metricHostList := range metricList.Metrics {
		for _, metricPlugin := range metricHostList.Plugins {
			metricTotalCount++
			rawMetrics, err := getMetrics(metricPlugin.PluginName, metricPlugin.PluginOption)
			if err != nil {
				return err
			} else if rawMetrics == "" {
				continue
			}
			metricData, timestamp, err := ParseMetricData(rawMetrics)
			if err != nil {
				return err
			}

			var metrics halib.MetricsData
			metrics.HostName = metricHostList.Hostname
			metrics.Timestamp = timestamp
			metrics.Metrics = metricData
			metricsDataBuffer = append(metricsDataBuffer, metrics)
		}
	}

	now := time.Now()
	err = SaveMetrics(now, metricsDataBuffer)
	return err
}

//SaveMetrics save metrics to dbms
func SaveMetrics(now time.Time, metricsData []halib.MetricsData) error {
	log := util.HappoAgentLogger()

	// Save Metrics
	transaction, err := db.DB.OpenTransaction()
	if err != nil {
		log.Error(err)
	}

	got, err := transaction.Get(
		[]byte(fmt.Sprintf("m-%d", now.Unix())),
		nil)
	if err != leveldbErrors.ErrNotFound {
		savedMetricsData := []halib.MetricsData{}
		dec := gob.NewDecoder(bytes.NewReader(got))
		dec.Decode(&savedMetricsData)
		metricsData = append(savedMetricsData, metricsData...)
	}

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err = enc.Encode(metricsData)
	if err != nil {
		log.Error(err)
	} else {
		transaction.Put(
			[]byte(fmt.Sprintf("m-%d", now.Unix())),
			b.Bytes(),
			nil)
	}

	err = transaction.Commit()
	if err != nil {
		//Fatal
		log.Fatalln(err)
	}

	// retire old metrics
	transaction, err = db.DB.OpenTransaction()
	if err != nil {
		log.Error(err)
	}
	oldestThreshold := now.Add(time.Duration(-1*db.MetricsMaxLifetimeSeconds) * time.Second)
	iter := transaction.NewIterator(
		&leveldbUtil.Range{
			Start: []byte("m-0"),
			Limit: []byte(fmt.Sprintf("m-%d", oldestThreshold.Unix()))},
		nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		transaction.Delete(key, nil)

		// logging
		unixTime, _ := strconv.Atoi(strings.SplitN(string(key), "-", 2)[1])
		expired := []halib.MetricsData{}
		dec := gob.NewDecoder(bytes.NewReader(value))
		dec.Decode(&expired)
		log.Warn("retire old metrics: key=%v(%v), value=%v\n", string(key), time.Unix(int64(unixTime), 0), expired)
	}
	iter.Release()

	err = transaction.Commit()
	if err != nil {
		log.Error(err)
	}

	return nil
}

// GetCollectedMetrics returns collected metrics. with no limit
func GetCollectedMetrics() []halib.MetricsData {
	return GetCollectedMetricsWithLimit(-1)
}

// GetCollectedMetricsWithLimit returns collected metrics. with max `limit`
func GetCollectedMetricsWithLimit(limit int) []halib.MetricsData {
	/*
		limit > 0 works fine. (otherwise, means unlimited)
	*/
	log := util.HappoAgentLogger()
	var collectedMetricsData []halib.MetricsData

	transaction, err := db.DB.OpenTransaction()
	if err != nil {
		log.Error(err)
	}

	var metricsData []halib.MetricsData
	var dec *gob.Decoder
	iter := transaction.NewIterator(
		leveldbUtil.BytesPrefix([]byte("m-")),
		nil)

	i := 0
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		metricsData = []halib.MetricsData{}
		dec = gob.NewDecoder(bytes.NewReader(value))
		err = dec.Decode(&metricsData)
		if err != nil {
			log.Error(err)
			continue
		}
		collectedMetricsData = append(collectedMetricsData, metricsData...)
		transaction.Delete(key, nil)

		i = i + 1
		if limit > 0 && i >= limit {
			break
		}
	}
	iter.Release()

	err = transaction.Commit()
	if err != nil {
		log.Error(err)
	}
	return collectedMetricsData
}

// getMetrics exec sensu plugin and get metrics
func getMetrics(pluginName string, pluginOption string) (string, error) {
	log := util.HappoAgentLogger()
	var plugin string

	for _, basePath := range strings.Split(SensuPluginPaths, ",") {
		plugin = path.Join(basePath, pluginName)
		_, err := os.Stat(plugin)
		if err == nil {
			if !util.Production {
				log.Debug(plugin)
			}
			break
		}
	}
	_, err := os.Stat(plugin)
	if err != nil {
		log.Error("Plugin not found:" + plugin)
		return "", nil
	}

	if !util.Production {
		log.Debug("Execute metric plugin:" + plugin)
	}
	exitstatus, stdout, _, err := util.ExecCommand(plugin, pluginOption)

	if err != nil {
		// timeout is onetime/runtime error, does not handle as serious error
		if timeoutError, ok := err.(*util.TimeoutError); ok {
			log.Errorf("Plugin timeout: %s %s", pluginName, timeoutError.Error())
			return stdout, nil
		}
		return "", err
	}
	if exitstatus != 0 {
		log.Error("Fail to get metrics:" + plugin)
		return "", nil
	}

	return stdout, nil
}

// ParseMetricData parse sensu-stype metrics output
func ParseMetricData(rawMetricdata string) (map[string]float64, int64, error) {
	var timestamp int64
	results := make(map[string]float64)

	for _, line := range strings.Split(rawMetricdata, "\n") {
		items := strings.Split(line, "\t")
		if len(items) != 3 {
			continue
		}
		value, err := strconv.ParseFloat(items[1], 64)
		if err != nil {
			return nil, 0, errors.New("Failed to parse values: " + line)
		}

		timestampValue, err := strconv.ParseInt(items[2], 10, 64)
		if err != nil {
			return nil, 0, errors.New("Failed to parse values: " + line)
		}
		if timestamp < timestampValue {
			timestamp = timestampValue
		}

		key := items[0]

		results[key] = value
	}

	return results, timestamp, nil
}

// GetMetricConfig returns required metrics from config file
func GetMetricConfig(configFile string) (halib.MetricConfig, error) {
	var metricConfig halib.MetricConfig

	buf, err := ioutil.ReadFile(configFile)
	if err != nil {
		return metricConfig, err
	}
	err = yaml.Unmarshal(buf, &metricConfig)
	if err != nil {
		return metricConfig, err
	}

	return metricConfig, nil
}

// SaveMetricConfig save metric config to config file
func SaveMetricConfig(config halib.MetricConfig, configFile string) error {
	buf, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(configFile, buf, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// GetMetricDataBufferStatus returns metric collection status
func GetMetricDataBufferStatus(extended bool) map[string]int64 {
	log := util.HappoAgentLogger()
	transaction, err := db.DB.OpenTransaction()
	if err != nil {
		log.Error(err)
		return map[string]int64{}
	}
	iter := transaction.NewIterator(
		leveldbUtil.BytesPrefix([]byte("m-")),
		nil)
	i := 0
	var firstKey, lastKey []byte

	if iter.First() {
		//have result
		firstKey = make([]byte, len(iter.Key()))
		copy(firstKey, iter.Key())
		i = i + 1
		if extended {
			// heavy... need long time (Disk IO bound)
			for iter.Next() {
				i = i + 1
			}
		}
		iter.Last()
		lastKey = make([]byte, len(iter.Key()))
		copy(lastKey, iter.Key())
	}
	iter.Release()
	transaction.Discard()

	length := i

	oldestTimestamp := int64(0)
	newestTimestamp := int64(0)
	if i > 0 {
		firstUnixTime, err := strconv.Atoi(strings.SplitN(string(firstKey), "-", 2)[1])
		if err != nil {
			log.Error(err)
		} else {
			oldestTimestamp = int64(firstUnixTime)
		}

		lastUnixTime, err := strconv.Atoi(strings.SplitN(string(lastKey), "-", 2)[1])
		if err != nil {
			log.Error(err)
		} else {
			newestTimestamp = int64(lastUnixTime)
		}
	}

	if extended {
		result := map[string]int64{
			"length":           int64(length),
			"oldest_timestamp": oldestTimestamp,
			"newest_timestamp": newestTimestamp,
		}
		return result
	}
	result := map[string]int64{
		"oldest_timestamp": oldestTimestamp,
		"newest_timestamp": newestTimestamp,
	}
	return result

}
