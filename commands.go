package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/heartbeatsjp/happo-agent/command"
	"github.com/heartbeatsjp/happo-lib"
)

var GlobalFlags = []cli.Flag{}

var daemonFlags = []cli.Flag{
	cli.IntFlag{
		Name:  "port, P",
		Value: happo_agent.DEFAULT_AGENT_PORT,
		Usage: "Listen port number",
	},
	cli.StringSliceFlag{
		Name:  "allowed-hosts, A",
		Value: &cli.StringSlice{},
		Usage: "Access allowed hosts (You can multiple define.)",
	},
	cli.StringFlag{
		Name:  "public-key, B",
		Value: happo_agent.TLS_PUBLIC_KEY,
		Usage: "TLS public key file path",
	},
	cli.StringFlag{
		Name:  "private-key, R",
		Value: happo_agent.TLS_PRIVATE_KEY,
		Usage: "TLS private key file path",
	},
	cli.StringFlag{
		Name:  "metric-config, M",
		Value: happo_agent.CONFIG_METRIC,
		Usage: "Metric config file path",
	},
	cli.StringFlag{
		Name:  "cpu-profile, C",
		Value: "",
		Usage: "CPU profile output.",
	},
	cli.IntFlag{
		Name:  "max-connections, X",
		Value: happo_agent.MAX_CONNECTIONS,
		Usage: "CPU profile output.",
	},
	cli.IntFlag{
		Name:  "command-timeout, T",
		Value: happo_agent.COMMAND_TIMEOUT,
		Usage: "Command execution timeout.",
	},
	cli.StringFlag{
		Name:  "logfile, l",
		Value: "happo-agent.log",
		Usage: "logfile.",
	},
}

var Commands = []cli.Command{
	{
		Name:   "daemon",
		Usage:  "Daemon mode (agent mode)",
		Action: command.CmdDaemon,
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
				Value: happo_agent.DEFAULT_AGENT_PORT,
				Usage: "Listen port number",
			},
			cli.StringFlag{
				Name:  "endpoint, e",
				Value: happo_agent.API_ENDPOINT,
				Usage: "Endpoint address",
			},
		},
	},
	{
		Name:   "is_added",
		Usage:  "Checking database who added the host.",
		Action: command.CmdIs_added,
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
				Value: happo_agent.DEFAULT_AGENT_PORT,
				Usage: "Listen port number",
			},
			cli.StringFlag{
				Name:  "endpoint, e",
				Value: happo_agent.API_ENDPOINT,
				Usage: "Endpoint address",
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
				Value: happo_agent.DEFAULT_AGENT_PORT,
				Usage: "Listen port number",
			},
			cli.StringFlag{
				Name:  "endpoint, e",
				Value: happo_agent.API_ENDPOINT,
				Usage: "Endpoint address",
			},
		},
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
