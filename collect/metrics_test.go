package collect

import (
	"os"
	"testing"
	"time"

	"github.com/heartbeatsjp/happo-agent/db"
	"github.com/heartbeatsjp/happo-agent/lib"

	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

const TestConfigFile = "./metrics_test.yaml"
const TestPlugin = "metrics_test_plugin"

var ConfigData = lib.MetricConfig{
	Metrics: []struct {
		Hostname string `yaml:"hostname" json:"Hostname"`
		Plugins  []struct {
			PluginName   string `yaml:"plugin_name" json:"Plugin_Name"`
			PluginOption string `yaml:"plugin_option" json:"Plugin_Option"`
		} `yaml:"plugins" json:"Plugins"`
	}{
		{
			Hostname: "localhost",
			Plugins: []struct {
				PluginName   string `yaml:"plugin_name" json:"Plugin_Name"`
				PluginOption string `yaml:"plugin_option" json:"Plugin_Option"`
			}{
				{
					PluginName:   "metrics_test_plugin",
					PluginOption: "",
				},
			},
		},
	},
}

func TestMetrics1(t *testing.T) {
	err := Metrics(TestConfigFile)
	assert.Nil(t, err)
}

func TestGetCollectedMetrics1(t *testing.T) {
	err := Metrics(TestConfigFile)
	assert.Nil(t, err)

	ret := GetCollectedMetrics()
	assert.NotNil(t, ret)
	assert.Nil(t, GetCollectedMetrics())
}

func TestGetCollectedMetricsWithLimit1(t *testing.T) {
	err := Metrics(TestConfigFile)
	assert.Nil(t, err)
	time.Sleep(1 * time.Second)
	err = Metrics(TestConfigFile)
	assert.Nil(t, err)

	ret := GetCollectedMetricsWithLimit(1)
	assert.NotNil(t, ret)
	assert.Equal(t, 1, len(ret))
	assert.NotNil(t, GetCollectedMetrics())
}

func TestGetMetrics1(t *testing.T) {
	ret, err := getMetrics(TestPlugin, "")
	assert.NotNil(t, ret)
	assert.Contains(t, ret, "usr.local.bin.metrics_test_plugin")
	assert.Nil(t, err)
}

func TestGetMetrics2(t *testing.T) {
	_, err := getMetrics("dummy", "")
	assert.Nil(t, err) // If plugin not found, not stop app.
}

func TestParseMetricData1(t *testing.T) {
	RetAssert := map[string]float64{"hoge": 10}

	ret, timestamp, err := ParseMetricData("hoge	10	1")
	assert.EqualValues(t, RetAssert, ret)
	assert.EqualValues(t, 1, timestamp)
	assert.Nil(t, err)
}

func TestParseMetricData2(t *testing.T) {
	ret, timestamp, err := ParseMetricData("hoge	foo	bar")
	assert.Nil(t, ret)
	assert.EqualValues(t, 0, timestamp)
	assert.NotNil(t, err)
}

func TestParseMetricData3(t *testing.T) {
	RetAssert := map[string]float64{}

	ret, timestamp, err := ParseMetricData("hoge")
	assert.EqualValues(t, RetAssert, ret)
	assert.EqualValues(t, 0, timestamp)
	assert.Nil(t, err)
}

func TestGetMetricConfig1(t *testing.T) {
	config, err := GetMetricConfig(TestConfigFile)
	assert.EqualValues(t, ConfigData, config)
	assert.Nil(t, err)
}

func TestGetMetricConfig2(t *testing.T) {
	_, err := GetMetricConfig("dummy")
	assert.NotNil(t, err)
}

func TestGetMetricConfig3(t *testing.T) {
	ret, err := GetMetricConfig("/proc/cpuinfo")
	assert.NotEqual(t, ConfigData, ret)
	assert.Nil(t, err)
}

func TestSaveMetricConfig1(t *testing.T) {
	err := SaveMetricConfig(ConfigData, TestConfigFile)
	assert.Nil(t, err)

	config, err := GetMetricConfig(TestConfigFile)
	assert.EqualValues(t, ConfigData, config)
	assert.Nil(t, err)
}

func TestSaveMetrics1(t *testing.T) {
	var err error
	metricsData1 := []lib.MetricsData{
		lib.MetricsData{HostName: "host1", Timestamp: 101, Metrics: map[string]float64{"val1": 111, "val2": 112}},
		lib.MetricsData{HostName: "host1", Timestamp: 102, Metrics: map[string]float64{"val1": 121, "val2": 122}},
	}
	metricsData2 := []lib.MetricsData{
		lib.MetricsData{HostName: "host2", Timestamp: 101, Metrics: map[string]float64{"val1": 211, "val2": 212}},
		lib.MetricsData{HostName: "host2", Timestamp: 102, Metrics: map[string]float64{"val1": 221, "val2": 222}},
	}

	err = SaveMetrics(time.Unix(1001, 0), metricsData1)
	assert.Nil(t, err)
	got := GetCollectedMetricsWithLimit(-1)
	assert.Equal(t, metricsData1, got)

	err = SaveMetrics(time.Unix(1002, 0), metricsData2)
	assert.Nil(t, err)
	got = GetCollectedMetricsWithLimit(-1)
	assert.Equal(t, metricsData2, got)
}

func TestSaveMetrics2(t *testing.T) {
	var err error
	metricsData1 := []lib.MetricsData{
		lib.MetricsData{HostName: "host1", Timestamp: 101, Metrics: map[string]float64{"val1": 111, "val2": 112}},
		lib.MetricsData{HostName: "host1", Timestamp: 102, Metrics: map[string]float64{"val1": 121, "val2": 122}},
	}
	metricsData2 := []lib.MetricsData{
		lib.MetricsData{HostName: "host2", Timestamp: 101, Metrics: map[string]float64{"val1": 211, "val2": 212}},
		lib.MetricsData{HostName: "host2", Timestamp: 102, Metrics: map[string]float64{"val1": 221, "val2": 222}},
	}

	err = SaveMetrics(time.Unix(1001, 0), metricsData1)
	assert.Nil(t, err)

	err = SaveMetrics(time.Unix(1002, 0), metricsData2)
	assert.Nil(t, err)

	got := GetCollectedMetricsWithLimit(-1)
	assert.Equal(t, append(metricsData1, metricsData2...), got)
}

func TestSaveMetrics3(t *testing.T) {
	var err error
	metricsData1 := []lib.MetricsData{
		lib.MetricsData{HostName: "host1", Timestamp: 101, Metrics: map[string]float64{"val1": 111, "val2": 112}},
		lib.MetricsData{HostName: "host1", Timestamp: 102, Metrics: map[string]float64{"val1": 121, "val2": 122}},
	}
	metricsData2 := []lib.MetricsData{
		lib.MetricsData{HostName: "host2", Timestamp: 201, Metrics: map[string]float64{"val1": 211, "val2": 212}},
		lib.MetricsData{HostName: "host2", Timestamp: 202, Metrics: map[string]float64{"val1": 221, "val2": 222}},
	}

	err = SaveMetrics(time.Unix(1000, 0), metricsData1)
	assert.Nil(t, err)

	err = SaveMetrics(time.Unix(1000, 0), metricsData2)
	assert.Nil(t, err)

	got := GetCollectedMetricsWithLimit(-1)
	assert.Equal(t, append(metricsData1, metricsData2...), got)
}

func TestSaveMetrics4(t *testing.T) {
	var err error
	metricsData1 := []lib.MetricsData{
		lib.MetricsData{HostName: "host1", Timestamp: 101, Metrics: map[string]float64{"val1": 111, "val2": 112}},
		lib.MetricsData{HostName: "host1", Timestamp: 102, Metrics: map[string]float64{"val1": 121, "val2": 122}},
	}
	metricsData2 := []lib.MetricsData{
		lib.MetricsData{HostName: "host2", Timestamp: 101, Metrics: map[string]float64{"val1": 211, "val2": 212}},
		lib.MetricsData{HostName: "host2", Timestamp: 102, Metrics: map[string]float64{"val1": 221, "val2": 222}},
	}

	err = SaveMetrics(time.Unix(1000, 0), metricsData1)
	assert.Nil(t, err)

	err = SaveMetrics(time.Unix(1000+db.MetricsMaxLifetimeSeconds+1, 0), metricsData2)
	assert.Nil(t, err)

	got := GetCollectedMetricsWithLimit(-1)
	//metricsData1 must expired
	assert.Equal(t, metricsData2, got)
}

func TestGetMetricDataBufferStatus1(t *testing.T) {
	var err error
	var savedMetricData map[string]int64
	metricsData1 := []lib.MetricsData{
		lib.MetricsData{HostName: "host1", Timestamp: 101, Metrics: map[string]float64{"val1": 111, "val2": 112}},
		lib.MetricsData{HostName: "host1", Timestamp: 102, Metrics: map[string]float64{"val1": 121, "val2": 122}},
	}
	metricsData2 := []lib.MetricsData{
		lib.MetricsData{HostName: "host2", Timestamp: 101, Metrics: map[string]float64{"val1": 211, "val2": 212}},
		lib.MetricsData{HostName: "host2", Timestamp: 102, Metrics: map[string]float64{"val1": 221, "val2": 222}},
	}

	//cleanup
	GetCollectedMetricsWithLimit(-1)

	savedMetricData = GetMetricDataBufferStatus()
	assert.Equal(t, int64(0), savedMetricData["length"])
	assert.Equal(t, int64(0), savedMetricData["capacity"])

	err = SaveMetrics(time.Unix(1000, 0), metricsData1)
	assert.Nil(t, err)

	savedMetricData = GetMetricDataBufferStatus()
	assert.Equal(t, int64(1), savedMetricData["length"])
	assert.Equal(t, int64(1), savedMetricData["capacity"])
	assert.Equal(t, int64(1000), savedMetricData["oldest_timestamp"])
	assert.Equal(t, int64(1000), savedMetricData["newest_timestamp"])

	err = SaveMetrics(time.Unix(1001, 0), metricsData2)
	assert.Nil(t, err)

	savedMetricData = GetMetricDataBufferStatus()
	assert.Equal(t, int64(2), savedMetricData["length"])
	assert.Equal(t, int64(2), savedMetricData["capacity"])
	assert.Equal(t, int64(1000), savedMetricData["oldest_timestamp"])
	assert.Equal(t, int64(1001), savedMetricData["newest_timestamp"])

	GetCollectedMetricsWithLimit(-1)

	savedMetricData = GetMetricDataBufferStatus()
	assert.Equal(t, int64(0), savedMetricData["length"])
	assert.Equal(t, int64(0), savedMetricData["capacity"])
}

func TestMain(m *testing.M) {
	//Mock
	DB, err := leveldb.Open(storage.NewMemStorage(), nil)
	if err != nil {
		os.Exit(1)
	}
	db.DB = DB
	os.Exit(m.Run())

	db.DB.Close()
}
