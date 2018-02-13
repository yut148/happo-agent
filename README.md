# happo-agent - yet another Nagios nrpe

[![wercker status](https://app.wercker.com/status/1d02bef8da5959d5b6456e25835ae026/s/ "wercker status")](https://app.wercker.com/project/byKey/1d02bef8da5959d5b6456e25835ae026)

## Description

`happo-agent` is yet another Nagios nrpe plugin. And improvement nrpe functions.

- More secure communication. Supports TLS 1.2.
- Less fork cost at bastion(proxy) mode. Proxy request handled by thread (not fork()).
- Metric collection. Compatible to Sensu plugin format.
- inventory collection.


## Usage

### Requires

- Red Hat Enterprise Linux (RHEL) 6.x, 7.x
- CentOS 6.x, 7.x
- Ubuntu 12.04 or later

### Daemon mode (for monitoring)

#### How to execute

```
/path/to/happo-agent daemon -A [Accept from IP/Subnet] -B [Public key file] -R [Private key file] -M [Metric config file (Accept empty file)]
```

**Many configuration can be with environment variables.**

See `/etc/default/happo-agent.env`
(example is in [contrib/etc/default/happo-agent.env](contrib/etc/default/happo-agent.env))

#### Monitoring

Call plugin from [`check_happo`](https://github.com/heartbeatsjp/check_happo), `happo-agent` calls local nagios plugin program. Then, return code and value to `check_happo`.

For more information, please see `check_happo` README.

#### Metric collection

Every one minute, execute sensu metrics plugin defined by `metrics.yaml`, and buffering results.

If you collect buffering results, you can use API `/metric` method.

#### Inventory collection

Get command based inventory data via API `/inventory` method.

### API client mode

You create `happo-agent` client management server if you want.

Use api client commands, `happo-agent` calls endpoint url which is client management server.

#### Host add request

```
/path/to/happo-agent add -e [ENDPOINT_URL] -g [GROUP_NAME[!SUB_GROUP_NAME]] -i [OWN_IP] -H [HOSTNAME] [-p BASTON_IP]
```

#### Is host available ?

```
/path/to/happo-agent is_added -e [ENDPOINT_URL] -g [GROUP_NAME[!SUB_GROUP_NAME]] -i [OWN_IP]
```

#### Host remove request

```
/path/to/happo-agent remove -e [ENDPOINT_URL] -g [GROUP_NAME[!SUB_GROUP_NAME]] -i [OWN_IP]
```

## Install

### Source based install (Use upstart)

```bash
$ sudo yum install epel-release
$ sudo yum install nagios-plugins-all
$ go get -dv github.com/heartbeatsjp/happo-agent
$ sudo install $GOHOME/src/bin/happo-agent /usr/local/bin/happo-agent
$ sudo install -d -m 755 /etc/happo
$ cd /etc/happo
$ sudo openssl genrsa -aes128 -out happo-agent.key 2048
$ sudo openssl req -new -key happo-agent.key -sha256 -out happo-agent.csr
$ sudo openssl x509 -in happo-agent.csr -days 3650 -req -signkey happo-agent.key -sha256 -out happo-agent.pub
$ sudo touch metrics.yaml
$ sudo chmod go-rwx happo-agent.key
$ sudo install contrib/etc/default/happo-agent.env /etc/default/happo-agent.env
$ sudo install contrib/etc/init/happo-agent.conf   /etc/init/happo-agent.conf
$ sudo initctl reload-configuration
$ sudo initctl start happo-agent
```

You want to use sensu metrics plugins, should install `/usr/local/bin`.

Pre build binary maybe useful.
[Releases · heartbeatsjp/happo\-agent](https://github.com/heartbeatsjp/happo-agent/releases)

### Metric collection configuration

metrics.yaml

```
metrics:
  - hostname: [HOSTNAME]
    plugins:
    - plugin_name: [Sensu plugin name (Path not needed)]
      plugin_option: [Sensu plugin name options]
    - ...
  - ...
```


## API

- Listen port: 6777 (Default)
- HTTPS, TLS 1.2, CipherSuite: `TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256`

### /

Check available.

- Input format
    - None
- Return format
    - String "OK"

```
$ wget -q --no-check-certificate -O - https://127.0.0.1:6777/
OK
```

### /proxy

Use agent bastion(proxy) mode.

- Input format
    - JSON
- Input variables
    - proxy\_hostport:
        - (Array) bastion_ip:port. It can multiple define.
    - request\_type: request type (e.g. `monitor`)
    - request\_json: Send JSON string to server.
- Return format
    - JSON
- Return variables
    - By `request_type` type.

In case `--proxy-timeout-seconds` reached, return `504 Gateway Timeout` .

```
$ wget -q --no-check-certificate -O - https://192.0.2.1:6777/proxy --post-data='{"proxy_hostport": ["198.51.100.1:6777"], "request_type": "monitor", "request_json": "{\"apikey\": \"\", \"plugin_name\": \"check_procs\", \"plugin_option\": \"-w 100 -c 200\"}"}'
{"return_value":1,"message":"PROCS WARNING: 168 processes\n"}
```

Example calls `wget host -> https://192.0.2.1:6777/proxy -> https://198.51.100.1:6777/monitor`.

### /inventory

Get inventory information from command.

- Input format
    - JSON
- Input variables
    - apikey: ""
    - command: execute command
    - command\_option: command option
- Return format
    - JSON
- Return variables
    - return\_code: commands return code
    - return\_value: commands return value (stdout, stderr)

```
$ wget -q --no-check-certificate -O - https://127.0.0.1:6777/inventory --post-data='{"apikey": "", "command": "uname", "command_option": "-a"}'
{"return_code":0,"return_value":"Linux saito-hb-vm101 2.6.32-573.3.1.el6.x86_64 #1 SMP Thu Aug 13 22:55:16 UTC 2015 x86_64 x86_64 x86_64 GNU/Linux\n"}
```

### /monitor

Call monitor plugin. It likes nrpe.

- Input format
    - JSON
- Input variables
    - apikey: ""
    - command: execute nagios plugin command
    - command\_option: command option
- Return format
    - JSON
- Return variables
    - return\_code: commands return code
    - return\_value: commands return value (stdout, stderr)

In case `--command-timeout` reached, return `500 Internal Server Error` .

```
$ wget -q --no-check-certificate -O - https://127.0.0.1:6777/monitor --post-data='{"apikey": "", "plugin_name": "check_procs", "plugin_option": "-w 100 -c 200"}'
{"return_value":1,"message":"PROCS WARNING: 168 processes\n"}
```

### /metric

Get collected metric values.

- Input format
    - JSON
- Input variables
    - apikey: ""
- Return format
    - JSON
- Return variables
    - MetricData:
        - (Array)
            - hostname: Hostname
            - timestamp: Unix time
            - metrics: metric name - metric value (key-value)
    - Message: message from agent (if error occurred)

```
$ wget -q --no-check-certificate -O - https://127.0.0.1:6777/metric --post-data='{"apikey": ""}'
{"metric_data":[{"hostname":"saito-hb-vm101","timestamp":1444028730,"metrics":{"linux.context_switches.context_switches":32662,"linux.disk.elapsed.iotime_sda":52,"linux.disk.elapsed.iotime_weighted_sda":82,"linux.disk.rwtime.tsreading_sda":0,"linux.disk.rwtime.tswriting_sda":82,"linux.forks.forks":88,"linux.interrupts.interrupts":19642,"linux.ss.CLOSE-WAIT":0,"linux.ss.CLOSING":0,"linux.ss.ESTAB":9,"linux.ss.FIN-WAIT-1":0,"linux.ss.FIN-WAIT-2":0,"linux.ss.LAST-ACK":0,"linux.ss.LISTEN":31,"linux.ss.SYN-RECV":0,"linux.ss.SYN-SENT":0,"linux.ss.TIME-WAIT":7,"linux.ss.UNCONN":0,"linux.ss.UNKNOWN":0,"linux.swap.pswpin":0,"linux.swap.pswpout":0,"linux.users.users":1}},…(snip)…],"message":""}
```

### /metric/append

Append metric values. (passive metrics collection)

- Input format
    - JSON
- Input variables
    - apikey: ""
    - MetricData:
        - (Array)
            - hostname: Hostname
            - timestamp: Unix time
            - metrics: metric name - metric value (key-value)
- Return format
    - JSON
- Return variables
    - Message: message from agent (if error occurred)

```
$ wget -q --no-check-certificate -O - https://127.0.0.1:6777/metric/append --post-data='{"apikey": "", "metric_data":[{"hostname":"saito-hb-vm101","timestamp":1444028730,"metrics":{"linux.context_switches.context_switches":32662,"linux.disk.elapsed.iotime_sda":52,"linux.disk.elapsed.iotime_weighted_sda":82,"linux.disk.rwtime.tsreading_sda":0,"linux.disk.rwtime.tswriting_sda":82,"linux.forks.forks":88,"linux.interrupts.interrupts":19642,"linux.ss.CLOSE-WAIT":0,"linux.ss.CLOSING":0,"linux.ss.ESTAB":9,"linux.ss.FIN-WAIT-1":0,"linux.ss.FIN-WAIT-2":0,"linux.ss.LAST-ACK":0,"linux.ss.LISTEN":31,"linux.ss.SYN-RECV":0,"linux.ss.SYN-SENT":0,"linux.ss.TIME-WAIT":7,"linux.ss.UNCONN":0,"linux.ss.UNKNOWN":0,"linux.swap.pswpin":0,"linux.swap.pswpout":0,"linux.users.users":1}},...(snip)...]}'
{"status": "ok", "message": ""}
```

### /metric/config/update

*TODO*

### /metric/status

replaced to /status

### /status

Get happo-agent status

- Input format
    - None
- Input variables
    - None
- Return format
    - JSON
- Return variables
    - app_version: happo-agent version ( equivalent to `happo-agent -v` )
    - uptime_seconds: seconds from happo-agent started
    - num_goroutine: number of goroutine
    - metric_buffer_status
        - oldest_timestamp: oldest Timestamp(int64) in metric_data_buffer
        - newest_timestamp: newest Timestamp(int64) in metric_data_buffer
    - callers: `filepath:linenum` of each goroutines

```
$ wget -q --no-check-certificate -O - https://127.0.0.1:6777/status
{"app_version":"1.0.0","uptime_seconds":13,"num_goroutine":15,"metric_buffer_status":{"newest_timestamp":1505180794,"oldest_timestamp":1504852118},"callers":["/goroot/src/runtime/extern.go:219","/gopath/src/github.com/heartbeatsjp/happo-agent/model/status.go:28",...(snip)...]}
```

### /status/memory

Get happo-agent memory usage status

- Input format
    - None
- Input variables
    - None
- Return format
    - JSON
- Return variables
    - runtime.MemStatus

```
$ wget -q --no-check-certificate -O - https://127.0.0.1:6777/status/memory
{"Alloc":7155296,"TotalAlloc":12148632,"Sys":14395640,"Lookups":34,"Mallocs":23456,"Frees":6565,...(snip)...}%
```

### /status/request

Get request status/count.

- Input format
    - None
- Input variables
    - None
- Return format
    - JSON
- Return variables
    - last1: Last 1 Minutes results
        - url: url
        - counts:
            - `<status_code>`
            - count
    - last5: Last 5 Minutes results
        - same as last1

```
$ wget -q --no-check-certificate -O - https://127.0.0.1:6777/status/request
{"last1":[{"url":"/","counts":{"200":3,"403":1}},{"url":"/proxy","counts":{"200":1,"403":1}}],"last5":[{"url":"/","counts":{"200":3,"403":1}},{"url":"/proxy","counts":{"200":1,"403":1}}]}
```

### /machine-state

Get machine state key list.

- Input format
    - None
- Input variables
    - None
- Return format
    - JSON
- Return variables
    - keys: machine-state key list

```
$ wget -q --no-check-certificate -O - https://127.0.0.1:6777/machine-state
{"keys":["2018-02-13T06:41:56Z"]}
```

### /machine-state/:key

Get machine state.

- Input format
    - None
- Input variables
    - key (can find from `/machine-state/` )
- Return format
    - JSON
- Return variables
    - machineState: command results

```
$ wget -q --no-check-certificate -O - "https://127.0.0.1:6777/machine-state/2018-02-13T06:41:56Z"
{"machineState":"********** w (2017-06-22T15:21:19+09:00) cron 15:21:19 up 13 days, ..."}
```

## DBMS

backend is boltdb.

[boltdb/bolt: An embedded key/value database for Go\.](https://github.com/boltdb/bolt)

- 2 buckets in one `data.db` file.
    - for Metric
    - for MachineState
- key is time(created at). Format is RFC3339
    - see [boltdb document](https://github.com/boltdb/bolt#range-scans)

## Contribution

1. Fork ([http://github.com/heartbeatsjp/happo-agent/fork](http://github.com/heartbeatsjp/happo-agent/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create a new Pull Request


## Author

- [Yuichiro Saito](https://github.com/koemu)
- [Toshiaki Baba](https://github.com/netmarkjp)

## License

Copyright 2016 HEARTBEATS Corporation.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
