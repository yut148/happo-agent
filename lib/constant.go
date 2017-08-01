package lib

// --- Constant values

// global

// DEFAULT_AGENT_PORT is default listen port of happo-agent
const DEFAULT_AGENT_PORT = 6777

// HTTP_TIMEOUT happo-agent http.Server ReadTimeout,WriteTimeout seconds
const HTTP_TIMEOUT = 60

// API_ENDPOINT is default API endpoint of happo backend
const API_ENDPOINT = "http://YOUR_MANAGEMENT_SERVER_HERE"

// for agent

// MAX_CONNECTIONS default happo-agent http.Server max connections
const MAX_CONNECTIONS = 1000

// COMMAND_KILLAFTER when command excess timeout, kill after COMMAND_KILLAFTER sec
const COMMAND_KILLAFTER = 3

// COMMAND_TIMEOUT command execution timeout(monitor, metric)
const COMMAND_TIMEOUT = 10

// ERROR_LOG_INTERVAL_SEC when monitor error(not MONITOR_OK), and ERROR_LOG_INTERVAL_SEC past from previous error, save sate snapshot.
const ERROR_LOG_INTERVAL_SEC = 600

// TLS_PRIVATE_KEY default TLS private key file path
const TLS_PRIVATE_KEY = "./happo-agent.key"

// TLS_PUBLIC_KEY default TLS public key file path
const TLS_PUBLIC_KEY = "./happo-agent.pub"

// for monitor

// MONITOR_OK is exit code OK (see also nagios plugin specification)
const MONITOR_OK = 0

// MONITOR_WARNING is exit code WARNING (see also nagios plugin specification)
const MONITOR_WARNING = 1

// MONITOR_ERROR is exit code ERROR(CRITICAL) (see also nagios plugin specification)
const MONITOR_ERROR = 2

// MONITOR_UNKNOWN is exit code UNKNOWN (see also nagios plugin specification)
const MONITOR_UNKNOWN = 3

// for metric

// NAGIOS_PLUGIN_PATHS is nagios plugin paths. many paths with comma
const NAGIOS_PLUGIN_PATHS = "/usr/local/hb-agent/bin,/usr/lib64/nagios/plugins,/usr/lib/nagios/plugins,/usr/local/nagios/libexec,/usr/local/bin"

// SENSU_PLUGIN_PATHS is sensu plugin paths. many paths with comma
const SENSU_PLUGIN_PATHS = "/usr/local/hb-agent/bin,/usr/local/bin"

// CONFIG_METRIC is default metric collection config path
const CONFIG_METRIC = "./metrics.yaml"
