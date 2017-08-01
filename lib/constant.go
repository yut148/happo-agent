package lib

// --- Constant values

// global
const DEFAULT_AGENT_PORT = 6777
const DEFAULT_MANAGER_PORT = 6776
const DEFAULT_PROXY_PORT = 8080
const HTTP_TIMEOUT = 60
const GROUP_NAME_DELIMITER = "!"
const LOCK_WAIT_SECONDS = 10
const API_ENDPOINT = "http://YOUR_MANAGEMENT_SERVER_HERE"

// for agent
const MAX_CONNECTIONS = 1000
const COMMAND_KILLAFTER = 3
const COMMAND_TIMEOUT = 10
const ERROR_LOG_INTERVAL_SEC = 600
const TLS_PRIVATE_KEY = "./happo-agent.key"
const TLS_PUBLIC_KEY = "./happo-agent.pub"

// for monitor
const MONITOR_OK = 0
const MONITOR_WARNING = 1
const MONITOR_ERROR = 2
const MONITOR_UNKNOWN = 3

// for metric
const NAGIOS_PLUGIN_PATHS = "/usr/local/hb-agent/bin,/usr/lib64/nagios/plugins,/usr/lib/nagios/plugins,/usr/local/nagios/libexec,/usr/local/bin"
const SENSU_PLUGIN_PATHS = "/usr/local/hb-agent/bin,/usr/local/bin"
const CONFIG_METRIC = "./metrics.yaml"
const CONFIG_HOSTLIST_FILENAME = "hostlist.yaml"
