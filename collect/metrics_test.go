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

const TEST_CONFIG_FILE = "./metrics_test.yaml"
const TEST_PLUGIN = "metrics_test_plugin"

var CONFIG_DATA = lib.MetricConfig{
	Metrics: []struct {
		Hostname string `yaml:"hostname"`
		Plugins  []struct {
			PluginName   string `yaml:"plugin_name"`
			PluginOption string `yaml:"plugin_option"`
		} `yaml:"plugins"`
	}{
		{
			Hostname: "localhost",
			Plugins: []struct {
				PluginName   string `yaml:"plugin_name"`
				PluginOption string `yaml:"plugin_option"`
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
	err := Metrics(TEST_CONFIG_FILE)
	assert.Nil(t, err)
}

func TestGetCollectedMetrics1(t *testing.T) {
	err := Metrics(TEST_CONFIG_FILE)
	assert.Nil(t, err)

	ret := GetCollectedMetrics()
	assert.NotNil(t, ret)
	assert.Nil(t, GetCollectedMetrics())
}

func TestGetCollectedMetricsWithLimit1(t *testing.T) {
	err := Metrics(TEST_CONFIG_FILE)
	assert.Nil(t, err)
	time.Sleep(1 * time.Second)
	err = Metrics(TEST_CONFIG_FILE)
	assert.Nil(t, err)

	ret := GetCollectedMetricsWithLimit(1)
	assert.NotNil(t, ret)
	assert.Equal(t, 1, len(ret))
	assert.NotNil(t, GetCollectedMetrics())
}

func TestGetMetrics1(t *testing.T) {
	ret, err := getMetrics(TEST_PLUGIN, "")
	assert.NotNil(t, ret)
	assert.Contains(t, ret, "usr.local.bin.metrics_test_plugin")
	assert.Nil(t, err)
}

func TestGetMetrics2(t *testing.T) {
	_, err := getMetrics("dummy", "")
	assert.Nil(t, err) // If plugin not found, not stop app.
}

func TestParseMetricData1(t *testing.T) {
	RET_ASSERT := map[string]float64{"hoge": 10}

	ret, timestamp, err := ParseMetricData("hoge	10	1")
	assert.EqualValues(t, ret, RET_ASSERT)
	assert.EqualValues(t, timestamp, 1)
	assert.Nil(t, err)
}

func TestParseMetricData2(t *testing.T) {
	ret, timestamp, err := ParseMetricData("hoge	foo	bar")
	assert.Nil(t, ret)
	assert.EqualValues(t, timestamp, 0)
	assert.NotNil(t, err)
}

func TestParseMetricData3(t *testing.T) {
	RET_ASSERT := map[string]float64{}

	ret, timestamp, err := ParseMetricData("hoge")
	assert.EqualValues(t, ret, RET_ASSERT)
	assert.EqualValues(t, timestamp, 0)
	assert.Nil(t, err)
}

func TestGetMetricConfig1(t *testing.T) {
	config, err := GetMetricConfig(TEST_CONFIG_FILE)
	assert.EqualValues(t, config, CONFIG_DATA)
	assert.Nil(t, err)
}

func TestGetMetricConfig2(t *testing.T) {
	_, err := GetMetricConfig("dummy")
	assert.NotNil(t, err)
}

func TestGetMetricConfig3(t *testing.T) {
	ret, err := GetMetricConfig("/proc/cpuinfo")
	assert.NotEqual(t, ret, CONFIG_DATA)
	assert.Nil(t, err)
}

func TestSaveMetricConfig1(t *testing.T) {
	err := SaveMetricConfig(CONFIG_DATA, TEST_CONFIG_FILE)
	assert.Nil(t, err)

	config, err := GetMetricConfig(TEST_CONFIG_FILE)
	assert.EqualValues(t, config, CONFIG_DATA)
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
	assert.Equal(t, got, metricsData1)

	err = SaveMetrics(time.Unix(1002, 0), metricsData2)
	assert.Nil(t, err)
	got = GetCollectedMetricsWithLimit(-1)
	assert.Equal(t, got, metricsData2)
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
	assert.Equal(t, got, append(metricsData1, metricsData2...))
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
	assert.Equal(t, got, append(metricsData1, metricsData2...))
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
	assert.Equal(t, got, metricsData2)
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
	assert.Equal(t, savedMetricData["length"], int64(0))
	assert.Equal(t, savedMetricData["capacity"], int64(0))

	err = SaveMetrics(time.Unix(1000, 0), metricsData1)
	assert.Nil(t, err)

	savedMetricData = GetMetricDataBufferStatus()
	assert.Equal(t, savedMetricData["length"], int64(1))
	assert.Equal(t, savedMetricData["capacity"], int64(1))
	assert.Equal(t, savedMetricData["oldest_timestamp"], int64(1000))
	assert.Equal(t, savedMetricData["newest_timestamp"], int64(1000))

	err = SaveMetrics(time.Unix(1001, 0), metricsData2)
	assert.Nil(t, err)

	savedMetricData = GetMetricDataBufferStatus()
	assert.Equal(t, savedMetricData["length"], int64(2))
	assert.Equal(t, savedMetricData["capacity"], int64(2))
	assert.Equal(t, savedMetricData["oldest_timestamp"], int64(1000))
	assert.Equal(t, savedMetricData["newest_timestamp"], int64(1001))

	GetCollectedMetricsWithLimit(-1)

	savedMetricData = GetMetricDataBufferStatus()
	assert.Equal(t, savedMetricData["length"], int64(0))
	assert.Equal(t, savedMetricData["capacity"], int64(0))
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
