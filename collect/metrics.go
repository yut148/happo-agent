package collect

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/heartbeatsjp/happo-agent/db"
	"github.com/heartbeatsjp/happo-agent/util"
	"github.com/heartbeatsjp/happo-lib"
	leveldbUtil "github.com/syndtr/goleveldb/leveldb/util"

	"gopkg.in/yaml.v2"
)

// --- Method

// メインクラス
func Metrics(config_path string) error {
	var metrics_data_buffer []happo_agent.MetricsData

	metric_list, err := GetMetricConfig(config_path)
	if err != nil {
		return err
	}

	metric_total_count := 0
	for _, metric_host_list := range metric_list.Metrics {
		for _, metric_plugin := range metric_host_list.Plugins {
			metric_total_count++
			raw_metrics, err := getMetrics(metric_plugin.Plugin_Name, metric_plugin.Plugin_Option)
			if err != nil {
				return err
			} else if raw_metrics == "" {
				continue
			}
			metric_data, timestamp, err := parseMetricData(raw_metrics)
			if err != nil {
				return err
			}

			var metrics happo_agent.MetricsData
			metrics.Host_Name = metric_host_list.Hostname
			metrics.Timestamp = timestamp
			metrics.Metrics = metric_data
			metrics_data_buffer = append(metrics_data_buffer, metrics)
		}
	}

	now := time.Now()

	// Save Metrics
	transaction, err := db.DB.OpenTransaction()
	if err != nil {
		log.Println(err)
	}

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err = enc.Encode(metrics_data_buffer)
	if err != nil {
		log.Println(err)
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
		log.Println(err)
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
		metricsData := []happo_agent.MetricsData{}
		dec := gob.NewDecoder(bytes.NewReader(value))
		dec.Decode(&metricsData)
		log.Printf("retire old metrics: key=%v(%v), value=%v\n", string(key), time.Unix(int64(unixTime), 0), metricsData)
	}
	iter.Release()

	err = transaction.Commit()
	if err != nil {
		log.Println(err)
	}

	return nil
}

// 取得済みのメトリックを返します
func GetCollectedMetrics() []happo_agent.MetricsData {
	var collectedMetricsData []happo_agent.MetricsData

	transaction, err := db.DB.OpenTransaction()
	if err != nil {
		log.Println(err)
	}

	var metricsData []happo_agent.MetricsData
	var dec *gob.Decoder
	iter := transaction.NewIterator(
		leveldbUtil.BytesPrefix([]byte("m-")),
		nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		metricsData = []happo_agent.MetricsData{}
		dec = gob.NewDecoder(bytes.NewReader(value))
		err = dec.Decode(&metricsData)
		if err != nil {
			log.Println(err)
			continue
		}
		collectedMetricsData = append(collectedMetricsData, metricsData...)
		transaction.Delete(key, nil)
	}
	iter.Release()

	err = transaction.Commit()
	if err != nil {
		log.Println(err)
	}
	return collectedMetricsData
}

// メトリック取得
func getMetrics(plugin_name string, plugin_option string) (string, error) {
	var plugin string

	for _, base_path := range strings.Split(happo_agent.SENSU_PLUGIN_PATHS, ",") {
		plugin = path.Join(base_path, plugin_name)
		_, err := os.Stat(plugin)
		if err == nil {
			if !util.Production {
				log.Println(plugin)
			}
			break
		}
	}
	_, err := os.Stat(plugin)
	if err != nil {
		log.Println("Plugin not found:" + plugin)
		return "", nil
	}

	if !util.Production {
		log.Println("Execute metric plugin:" + plugin)
	}
	exitstatus, stdout, _, err := util.ExecCommand(plugin, plugin_option)

	if err != nil {
		return "", err
	}
	if exitstatus != 0 {
		log.Println("Fail to get metrics:" + plugin)
		return "", nil
	}

	return stdout, nil
}

// Sensu形式のメトリックをパースします
func parseMetricData(raw_metricdata string) (map[string]float64, int64, error) {
	var timestamp int64 = 0
	results := make(map[string]float64)

	for _, line := range strings.Split(raw_metricdata, "\n") {
		items := strings.Split(line, "\t")
		if len(items) != 3 {
			continue
		}
		value, err := strconv.ParseFloat(items[1], 64)
		if err != nil {
			return nil, 0, errors.New("Failed to parse values: " + line)
		}

		timestamp_value, err := strconv.ParseInt(items[2], 10, 64)
		if err != nil {
			return nil, 0, errors.New("Failed to parse values: " + line)
		}
		if timestamp < timestamp_value {
			timestamp = timestamp_value
		}

		key := items[0]

		results[key] = value
	}

	return results, timestamp, nil
}

// 取得するメトリックをリストアップします
func GetMetricConfig(config_file string) (happo_agent.MetricConfig, error) {
	var metric_config happo_agent.MetricConfig

	buf, err := ioutil.ReadFile(config_file)
	if err != nil {
		return metric_config, err
	}
	err = yaml.Unmarshal(buf, &metric_config)
	if err != nil {
		return metric_config, err
	}

	return metric_config, nil
}

// 取得したいメトリックの情報を保存します
func SaveMetricConfig(config happo_agent.MetricConfig, config_file string) error {
	buf, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(config_file, buf, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func GetMetricDataBufferStatus() map[string]int64 {

	transaction, err := db.DB.OpenTransaction()
	if err != nil {
		log.Println(err)
		return map[string]int64{}
	}
	iter := transaction.NewIterator(
		leveldbUtil.BytesPrefix([]byte("m-")),
		nil)
	i := 0
	var firstKey, lastKey []byte
	for iter.Next() {
		if i == 0 {
			firstKey = make([]byte, len(iter.Key()))
			copy(firstKey, iter.Key())
		}
		lastKey = iter.Key()
		i = i + 1
	}
	iter.Release()
	transaction.Discard()

	length := i
	capacity := i

	oldest_timestamp := int64(0)
	newest_timestamp := int64(0)
	if i > 0 {
		firstUnixTime, err := strconv.Atoi(strings.SplitN(string(firstKey), "-", 2)[1])
		if err != nil {
			log.Println(err)
		} else {
			oldest_timestamp = int64(firstUnixTime)
		}

		lastUnixTime, err := strconv.Atoi(strings.SplitN(string(lastKey), "-", 2)[1])
		if err != nil {
			log.Println(err)
		} else {
			newest_timestamp = int64(lastUnixTime)
		}
	}

	result := map[string]int64{
		"length":           int64(length),
		"capacity":         int64(capacity),
		"oldest_timestamp": oldest_timestamp,
		"newest_timestamp": newest_timestamp,
	}

	return result
}
