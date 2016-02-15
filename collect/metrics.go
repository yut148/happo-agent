package collect

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/heartbeatsjp/happo-agent/util"
	"github.com/heartbeatsjp/happo-lib"

	"gopkg.in/yaml.v2"
)

// --- Package Variables
var metrics_data_buffer []happo_agent.MetricsData
var metrics_data_buffer_mutex = sync.Mutex{}

// --- Method

// メインクラス
func Metrics(config_path string) error {

	metric_list, err := GetMetricConfig(config_path)
	if err != nil {
		return err
	}

	metrics_data_buffer_mutex.Lock()
	defer metrics_data_buffer_mutex.Unlock()

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

	// fuzzy capacity control. FIXME
	// keep about 1 week. (60 times/hour * 24 times/day * 7 days/week)
	if metric_total_count > 0 && len(metrics_data_buffer) > metric_total_count*10080 {
		metrics_data_buffer = metrics_data_buffer[metric_total_count:]
	}

	return nil
}

// 取得済みのメトリックを返します
func GetCollectedMetrics() []happo_agent.MetricsData {
	metrics_data_buffer_mutex.Lock()
	defer metrics_data_buffer_mutex.Unlock()

	collected_metrics_data := make([]happo_agent.MetricsData, len(metrics_data_buffer))

	copy(collected_metrics_data, metrics_data_buffer)
	metrics_data_buffer = nil // バッファはオールクリア

	return collected_metrics_data
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
	metrics_data_buffer_mutex.Lock()
	defer metrics_data_buffer_mutex.Unlock()

	length := len(metrics_data_buffer)
	capacity := cap(metrics_data_buffer)

	oldest_timestamp := int64(0)
	newest_timestamp := int64(0)
	if len(metrics_data_buffer) > 0 {
		oldest_timestamp = metrics_data_buffer[0].Timestamp
		newest_timestamp = metrics_data_buffer[len(metrics_data_buffer)-1].Timestamp
	}

	result := map[string]int64{
		"length":           int64(length),
		"capacity":         int64(capacity),
		"oldest_timestamp": oldest_timestamp,
		"newest_timestamp": newest_timestamp,
	}

	return result
}
