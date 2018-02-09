package halib

// --- Struct

// MetricsData is actual metrics
type MetricsData struct {
	HostName  string             `json:"hostname"`
	Timestamp int64              `json:"timestamp"`
	Metrics   map[string]float64 `json:"metrics"`
}

// InventoryData is actual inventory
type InventoryData struct {
	GroupName     string `json:"group_name"`
	IP            string `json:"ip"`
	Command       string `json:"command"`
	CommandOption string `json:"command_option"`
	ReturnCode    int    `json:"return_code"`
	ReturnValue   string `json:"return_value"`
	Created       string `json:"created"`
}

// --- Request Parameter

// ProxyRequest is /proxy API
type ProxyRequest struct {
	ProxyHostPort []string `json:"proxy_hostport"`
	RequestType   string   `json:"request_type"`
	RequestJSON   []byte   `json:"request_json"`
}

// MonitorRequest is /monitor API
type MonitorRequest struct {
	APIKey       string `json:"apikey"`
	PluginName   string `json:"plugin_name"  binding:"required"`
	PluginOption string `json:"plugin_option"`
}

// MetricRequest is /metric API
type MetricRequest struct {
	APIKey string `json:"apikey"`
}

// MetricAppendRequest is /metric/append API
type MetricAppendRequest struct {
	APIKey     string        `json:"apikey"`
	MetricData []MetricsData `json:"metric_data"`
}

// MetricConfigUpdateRequest is /metric/config/update API
type MetricConfigUpdateRequest struct {
	APIKey string       `json:"apikey"`
	Config MetricConfig `json:"config"`
}

// InventoryRequest is /inventory API
type InventoryRequest struct {
	APIKey        string `json:"apikey"`
	Command       string `json:"command"`
	CommandOption string `json:"command_option"`
}

// ManageRequest is Manage API
type ManageRequest struct {
	APIKey   string           `json:"apikey"`
	Hostdata CrawlConfigAgent `json:"hostdata"`
}

// --- Response Parameter

// MonitorResponse is /monitor API
type MonitorResponse struct {
	ReturnValue int    `json:"return_value"`
	Message     string `json:"message"`
}

// MetricResponse is /metric API
type MetricResponse struct {
	MetricData []MetricsData `json:"metric_data"`
	Message    string        `json:"message"`
}

// MetricAppendResponse is /metric/append API
type MetricAppendResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// MetricConfigUpdateResponse is /metric/config/update API
type MetricConfigUpdateResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// InventoryResponse is /inventory API
type InventoryResponse struct {
	ReturnCode  int    `json:"return_code"`
	ReturnValue string `json:"return_value"`
}

// ManageResponse is Manage API
type ManageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// StatusResponse is /status API
type StatusResponse struct {
	AppVersion         string            `json:"app_version"`
	UptimeSeconds      int64             `json:"uptime_seconds"`
	NumGoroutine       int               `json:"num_goroutine"`
	MetricBufferStatus map[string]int64  `json:"metric_buffer_status"`
	Callers            []string          `json:"callers"`
	LevelDBProperties  map[string]string `json:"leveldb_properties"`
}

// RequestStatusResponse is /status/request API
type RequestStatusResponse struct {
	Last1 []RequestStatusData `json:"last1"`
	Last5 []RequestStatusData `json:"last5"`
}

// RequestStatusData is data part of RequestStatusResponse
type RequestStatusData struct {
	URL    string         `json:"url"`
	Counts map[int]uint64 `json:"counts"`
}
