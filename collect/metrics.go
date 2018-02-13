package collect

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/heartbeatsjp/happo-agent/db"
	"github.com/heartbeatsjp/happo-agent/halib"
	"github.com/heartbeatsjp/happo-agent/util"

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
	var err error
	log := util.HappoAgentLogger()

	// Save Metrics
	err = db.DB.Update(func(tx *bolt.Tx) error {
		bucket := db.MetricBucket(tx)
		got := bucket.Get(db.TimeToKey(now))
		if len(got) > 0 {
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
			bucket.Put(
				db.TimeToKey(now),
				b.Bytes())
		}

		return nil
	})
	if err != nil {
		//Fatal
		log.Fatalln("commit transaction failed at SaveMetrics.", err)
	}

	// retire old metrics
	err = db.DB.Update(func(tx *bolt.Tx) error {
		bucket := db.MetricBucket(tx)
		cursor := bucket.Cursor()
		oldestThreshold := now.Add(time.Duration(-1*db.MetricsMaxLifetimeSeconds) * time.Second)

		retrivedKeys := make([][]byte, 0)
		oldestThresholdKey := db.TimeToKey(oldestThreshold)
		for key, value := cursor.First(); key != nil && bytes.Compare(key, oldestThresholdKey) <= 0; key, value = cursor.Next() {
			retrivedKeys = append(retrivedKeys, key)

			// logging
			expired := []halib.MetricsData{}
			dec := gob.NewDecoder(bytes.NewReader(value))
			dec.Decode(&expired)
			log.Warn("retire old metrics: key=%v, value=%v\n", string(key), expired)
		}

		for _, key := range retrivedKeys {
			bucket.Delete(key)
		}
		return nil
	})
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
	var err error
	log := util.HappoAgentLogger()
	var collectedMetricsData []halib.MetricsData

	err = db.DB.Update(func(tx *bolt.Tx) error {
		var metricsData []halib.MetricsData
		var dec *gob.Decoder

		bucket := db.MetricBucket(tx)
		cursor := bucket.Cursor()

		retrivedKeys := make([][]byte, 0)

		i := 0
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			metricsData = []halib.MetricsData{}
			dec = gob.NewDecoder(bytes.NewReader(value))
			err = dec.Decode(&metricsData)
			if err != nil {
				log.Error(err)
				continue
			}
			collectedMetricsData = append(collectedMetricsData, metricsData...)
			retrivedKeys = append(retrivedKeys, key)

			i = i + 1
			if limit > 0 && i >= limit {
				break
			}
		}

		for _, key := range retrivedKeys {
			bucket.Delete(key)
		}
		return nil
	})
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

	var firstKey, lastKey []byte
	var length int
	err := db.DB.View(func(tx *bolt.Tx) error {
		bucket := db.MetricBucket(tx)
		cursor := bucket.Cursor()

		//have result
		firstKey, _ = cursor.First()
		length = bucket.Stats().KeyN

		lastKey, _ = cursor.Last()
		return nil
	})
	if err != nil {
		log.Error(err)
		return map[string]int64{}
	}

	oldestTimestamp := int64(0)
	newestTimestamp := int64(0)
	if length > 0 {
		oldestTimestamp = db.KeyToUnixtime(firstKey)
		newestTimestamp = db.KeyToUnixtime(lastKey)
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
