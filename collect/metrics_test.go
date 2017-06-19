package collect

import (
	"os"
	"testing"
	"time"

	"github.com/heartbeatsjp/happo-agent/db"
	"github.com/heartbeatsjp/happo-lib"

	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

const TEST_CONFIG_FILE = "./metrics_test.yaml"
const TEST_PLUGIN = "metrics_test_plugin"

var CONFIG_DATA = happo_agent.MetricConfig{
	Metrics: []struct {
		Hostname string `yaml:"hostname"`
		Plugins  []struct {
			Plugin_Name   string `yaml:"plugin_name"`
			Plugin_Option string `yaml:"plugin_option"`
		} `yaml:"plugins"`
	}{
		{
			Hostname: "localhost",
			Plugins: []struct {
				Plugin_Name   string `yaml:"plugin_name"`
				Plugin_Option string `yaml:"plugin_option"`
			}{
				{
					Plugin_Name:   "metrics_test_plugin",
					Plugin_Option: "",
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

	ret, timestamp, err := parseMetricData("hoge	10	1")
	assert.EqualValues(t, ret, RET_ASSERT)
	assert.EqualValues(t, timestamp, 1)
	assert.Nil(t, err)
}

func TestParseMetricData2(t *testing.T) {
	ret, timestamp, err := parseMetricData("hoge	foo	bar")
	assert.Nil(t, ret)
	assert.EqualValues(t, timestamp, 0)
	assert.NotNil(t, err)
}

func TestParseMetricData3(t *testing.T) {
	RET_ASSERT := map[string]float64{}

	ret, timestamp, err := parseMetricData("hoge")
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
