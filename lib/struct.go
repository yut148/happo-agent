package lib

// --- Struct

// metric buffer
type MetricsData struct {
	HostName  string             `json:"hostname"`
	Timestamp int64              `json:"timestamp"`
	Metrics   map[string]float64 `json:"metrics"`
}

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

// /proxy
type ProxyRequest struct {
	ProxyHostPort []string `json:"proxy_hostport"`
	RequestType   string   `json:"request_type"`
	RequestJSON   []byte   `json:"request_json"`
}

// /monitor API
type MonitorRequest struct {
	APIKey       string `json:"apikey"`
	PluginName   string `json:"plugin_name"  binding:"required"`
	PluginOption string `json:"plugin_option"`
}

// /metric API
type MetricRequest struct {
	APIKey string `json:"apikey"`
}

// /metric/append API
type MetricAppendRequest struct {
	APIKey     string        `json:"apikey"`
	MetricData []MetricsData `json:"metric_data"`
}

// /metric/config/update API
type MetricConfigUpdateRequest struct {
	APIKey string       `json:"apikey"`
	Config MetricConfig `json:"config"`
}

// /inventory API
type InventoryRequest struct {
	APIKey        string `json:"apikey"`
	Command       string `json:"command"`
	CommandOption string `json:"command_option"`
}

// Manage API
type ManageRequest struct {
	APIKey   string           `json:"apikey"`
	Hostdata CrawlConfigAgent `json:"hostdata"`
}

// --- Response Parameter

// /monitor API
type MonitorResponse struct {
	ReturnValue int    `json:"return_value"`
	Message     string `json:"message"`
}

// /metric API
type MetricResponse struct {
	MetricData []MetricsData `json:"metric_data"`
	Message    string        `json:"message"`
}

// /metric/append API
type MetricAppendResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// /metric/config/update API
type MetricConfigUpdateResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// /inventory API
type InventoryResponse struct {
	ReturnCode  int    `json:"return_code"`
	ReturnValue string `json:"return_value"`
}

// Manage API
type ManageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
