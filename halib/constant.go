package halib

// --- Constant values

// global

// DefaultAgentPort is default listen port of happo-agent
const DefaultAgentPort = 6777

// DefaultServerHTTPTimeout happo-agent http.Server ReadTimeout,WriteTimeout seconds
const DefaultServerHTTPTimeout = 60

// DefaultAPIEndpoint is default API endpoint of happo backend
const DefaultAPIEndpoint = "http://YOUR_MANAGEMENT_SERVER_HERE"

// for agent

// DefaultServerMaxConnections default happo-agent http.Server max connections
const DefaultServerMaxConnections = 1000

// CommandKillAfterSeconds when command excess timeout, kill after COMMAND_KILLAFTER sec
const CommandKillAfterSeconds = 3

// DefaultCommandTimeout command execution timeout(monitor, metric)
const DefaultCommandTimeout = 10

// ErrorLogIntervalSeconds when monitor error(not MonitorOK), and ErrorLogIntervalSeconds past from previous error, save sate snapshot.
const DefaultErrorLogIntervalSeconds = 600

// DefaultTLSPrivateKey default TLS private key file path
const DefaultTLSPrivateKey = "./happo-agent.key"

// DefaultTLSPublicKey default TLS public key file path
const DefaultTLSPublicKey = "./happo-agent.pub"

// for monitor

// MonitorOK is exit code OK (see also nagios plugin specification)
const MonitorOK = 0

// MonitorWarning is exit code WARNING (see also nagios plugin specification)
const MonitorWarning = 1

// MonitorError is exit code ERROR(CRITICAL) (see also nagios plugin specification)
const MonitorError = 2

// MonitorUnknown is exit code UNKNOWN (see also nagios plugin specification)
const MonitorUnknown = 3

// for metric

// DefaultNagiosPluginPaths is nagios plugin paths. many paths with comma
const DefaultNagiosPluginPaths = "/usr/local/hb-agent/bin,/usr/lib64/nagios/plugins,/usr/lib/nagios/plugins,/usr/local/nagios/libexec,/usr/local/bin"

// DefaultSensuPluginPaths is sensu plugin paths. many paths with comma
const DefaultSensuPluginPaths = "/usr/local/hb-agent/bin,/usr/local/bin"

// DefaultMetricsConfigPath is default metric collection config path
const DefaultMetricsConfigPath = "./metrics.yaml"
