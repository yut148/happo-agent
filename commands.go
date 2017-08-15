package main

import (
	"fmt"
	"os"

	_ "net/http/pprof"

	"github.com/codegangsta/cli"
	"github.com/heartbeatsjp/happo-agent/command"
	"github.com/heartbeatsjp/happo-agent/db"
	"github.com/heartbeatsjp/happo-agent/halib"
)

// GlobalFlags are global level options
var GlobalFlags = []cli.Flag{}

var daemonFlags = []cli.Flag{
	cli.IntFlag{
		Name:  "port, P",
		Value: halib.DefaultAgentPort,
		Usage: "Listen port number",
	},
	cli.StringSliceFlag{
		Name:  "allowed-hosts, A",
		Value: &cli.StringSlice{},
		Usage: "Access allowed hosts (You can multiple define.)",
	},
	cli.StringFlag{
		Name:  "public-key, B",
		Value: halib.DefaultTLSPublicKey,
		Usage: "TLS public key file path",
	},
	cli.StringFlag{
		Name:  "private-key, R",
		Value: halib.DefaultTLSPrivateKey,
		Usage: "TLS private key file path",
	},
	cli.StringFlag{
		Name:  "metric-config, M",
		Value: halib.DefaultMetricsConfigPath,
		Usage: "Metric config file path",
	},
	cli.StringFlag{
		Name:  "cpu-profile, C",
		Value: "",
		Usage: "CPU profile output.",
	},
	cli.IntFlag{
		Name:  "max-connections, X",
		Value: halib.DefaultServerMaxConnections,
		Usage: "CPU profile output.",
	},
	cli.IntFlag{
		Name:  "command-timeout, T",
		Value: halib.DefaultCommandTimeout,
		Usage: "Command execution timeout.",
	},
	cli.StringFlag{
		Name:  "logfile, l",
		Value: "happo-agent.log",
		Usage: "logfile.",
	},
	cli.StringFlag{
		Name:   "dbfile, d",
		Value:  "happo-agent.db",
		Usage:  "dbfile",
		EnvVar: "HAPPO_AGENT_DBFILE",
	},
	cli.Int64Flag{
		Name:   "metrics-max-lifetime-seconds",
		Value:  db.MetricsMaxLifetimeSeconds,
		Usage:  "Metrics Max Lifetime Seconds.",
		EnvVar: "HAPPO_AGENT_METRICS_MAX_LIFETIME_SECONDS",
	},
	cli.Int64Flag{
		Name:   "machine-state-max-lifetime-seconds",
		Value:  db.MachineStateMaxLifetimeSeconds,
		Usage:  "Machine State Max Lifetime Seconds.",
		EnvVar: "HAPPO_AGENT_MACHINE_STATE_MAX_LIFETIME_SECONDS",
	},
	cli.Int64Flag{
		Name:   "proxy-timeout-seconds",
		Value:  180,
		Usage:  "/proxy timeout Seconds.",
		EnvVar: "HAPPO_AGENT_PROXY_TIMEOUT_SECONDS",
	},
	cli.Int64Flag{
		Name:   "error-log-interval-seconds",
		Value:  halib.DefaultErrorLogIntervalSeconds,
		Usage:  "Error log collection interval Seconds(when >0, disable error log collection).",
		EnvVar: "HAPPO_AGENT_ERROR_LOG_INTERVAL_SECONDS",
	},
	cli.StringFlag{
		Name:   "nagios-plugin-paths",
		Value:  halib.DefaultNagiosPluginPaths,
		Usage:  "nagios-plugin paths.",
		EnvVar: "HAPPO_AGENT_NAGIOS_PLUGIN_PATHS",
	},
	cli.StringFlag{
		Name:   "sensu-plugin-paths",
		Value:  halib.DefaultSensuPluginPaths,
		Usage:  "sensu-plugin paths.",
		EnvVar: "HAPPO_AGENT_SENSU_PLUGIN_PATHS",
	},
}

// Commands is list of subcommand
var Commands = []cli.Command{
	{
		Name:   "_daemon",
		Usage:  "Daemon mode (agent mode)",
		Action: command.CmdDaemon,
		Flags:  daemonFlags,
	},
	{
		Name:   "daemon",
		Usage:  "Daemon mode (agent mode)",
		Action: command.CmdDaemonWrapper,
		Flags:  daemonFlags,
	},
	{
		Name:   "add",
		Usage:  "Add to metric server",
		Action: command.CmdAdd,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "group_name, g",
				Usage: "Group name",
			},
			cli.StringFlag{
				Name:  "ip, i",
				Usage: "IP address (This host!)",
			},
			cli.StringFlag{
				Name:  "hostname, H",
				Usage: "Hostname (This host!)",
			},
			cli.StringSliceFlag{
				Name:  "proxy, p",
				Value: &cli.StringSlice{},
				Usage: "Proxy host ip:port (You can multiple define.)",
			},
			cli.IntFlag{
				Name:  "port, P",
				Value: halib.DefaultAgentPort,
				Usage: "Listen port number",
			},
			cli.StringFlag{
				Name:   "endpoint, e",
				Value:  halib.DefaultAPIEndpoint,
				Usage:  "API Endpoint address",
				EnvVar: "HAPPO_AGENT_API_ENDPOINT",
			},
		},
	},
	{
		Name:   "is_added",
		Usage:  "Checking database who added the host.",
		Action: command.CmdIsAdded,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "group_name, g",
				Usage: "Group name",
			},
			cli.StringFlag{
				Name:  "ip, i",
				Usage: "IP address (This host!)",
			},
			cli.IntFlag{
				Name:  "port, P",
				Value: halib.DefaultAgentPort,
				Usage: "Listen port number",
			},
			cli.StringFlag{
				Name:   "endpoint, e",
				Value:  halib.DefaultAPIEndpoint,
				Usage:  "API Endpoint address",
				EnvVar: "HAPPO_AGENT_API_ENDPOINT",
			},
		},
	},
	{
		Name:   "remove",
		Usage:  "Remove host",
		Action: command.CmdRemove,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "group_name, g",
				Usage: "Group name",
			},
			cli.StringFlag{
				Name:  "ip, i",
				Usage: "IP address (This host!)",
			},
			cli.IntFlag{
				Name:  "port, P",
				Value: halib.DefaultAgentPort,
				Usage: "Listen port number",
			},
			cli.StringFlag{
				Name:   "endpoint, e",
				Value:  halib.DefaultAPIEndpoint,
				Usage:  "API Endpoint address",
				EnvVar: "HAPPO_AGENT_API_ENDPOINT",
			},
		},
	},
	{
		Name:   "append_metric",
		Usage:  "Append Metric.",
		Action: command.CmdAppendMetric,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "hostname, H",
				Usage:  "Hostname",
				EnvVar: "HAPPO_AGENT_HOSTNAME",
			},
			cli.StringFlag{
				Name:   "bastion-endpoint, b",
				Value:  "https://127.0.0.1:6777",
				Usage:  "Bastion (Nearby happo-agent) endpoint address",
				EnvVar: "HAPPO_AGENT_BASTION_ENDPOINT",
			},
			cli.StringFlag{
				Name:   "datafile",
				Value:  "-",
				Usage:  "sensu format datafile(default: - (stdin))",
				EnvVar: "HAPPO_AGENT_DATAFILE",
			},
			cli.StringFlag{
				Name:   "api-key, a",
				Value:  "",
				Usage:  "API Key",
				EnvVar: "HAPPO_AGENT_API_KEY",
			},
			cli.BoolFlag{
				Name:   "dry-run, n",
				Usage:  "dry run(NOT post to bastion)",
				EnvVar: "HAPPO_AGENT_DRY_RUN",
			},
		},
	},
}

// CommandNotFound implements action when subcommand not found
func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
