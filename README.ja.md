# happo-agent README


## Description

ネットワーク越しにサーバを監視できるエージェントプログラムです。

基本的な機能はNagios nrpeに近いのですが、以下の優位性があります。

- よりセキュアな通信を担保すべく、通信の暗号化時にTLS 1.2で提供している方式を利用します。nrpeではDHであり、2015年現在では弱めの方式が採用されています。
- 踏み台として動作する際の負担を減らします。リクエスト単位でスレッドが生成され、監視対象サーバへ中継を行います。nrpeではリクエスト単位で更に`check_nrpe`コマンドが`fork`され、中継コストが高くなります。これは、監視対象サーバの数が多くなるほどパフォーマンスに対し顕著に現れます。
- Sensu互換のプラグインを利用し、メトリックを取得することができます。また、メトリックの収集とメトリックサーバへの転送は独立しています。従って、エージェントが正常に動作し続けていれば仮にメトリックサーバがダウンしてもメトリックの取得漏れが発生することはありません。この機能はnrpeにはありません。
- インベントリ情報取得機能があります。リモートにある監視対象のサーバのリクエストに応じて、必要な情報を取得することができます。この機能はnrpeにはありません。


## Usage

### Requires

動作確認を行っているのは、以下のOSです。

- Red Hat Enterprise Linux (RHEL) 6.x, 7.x
- CentOS 6.x, 7.x
- Ubuntu 12.04 or later

### デーモンモード

#### 起動方法

以下の通りコマンドを実行すると、エージェントがデーモンとして起動します。

```
/path/to/happo-agent daemon -A [Accept from IP/Subnet] -B [Public key file] -R [Private key file] -M [Metric config file (Accept empty file)]
```

#### 監視

`check_happo`コマンドを用いてエージェントを呼び出します。エージェントは、指定されたNagios形式のプラグインを実行し、`check_happo`に返します。

呼び出し方は `check_happo`の説明をご覧ください。

#### メトリック取得

`metrics.yaml`ファイルに定義されたコマンドを1分に1回実行し、エージェントに蓄積します。

蓄積したデータはAPIを用いて取得します。詳細はAPI `/metric`の説明をご覧ください。

#### インベントリ情報取得

エージェントがインストールされたサーバのインベントリ情報をAPIを用いて取得します。詳細はAPI `inventory`の説明をご覧ください。

### APIクライアントモード

#### ホスト登録

このエージェントの存在を通知します。

```
/path/to/happo-agent add -e [ENDPOINT_URL] -g [GROUP_NAME[!SUB_GROUP_NAME]] -i [OWN_IP] -H [HOSTNAME] [-p BASTON_IP]
```

#### 登録済みかの確認

このエージェントが登録されているかを確認します。

```
/path/to/happo-agent remove -e [ENDPOINT_URL] -g [GROUP_NAME[!SUB_GROUP_NAME]] -i [OWN_IP]
```

#### ホスト削除

ホスト一覧から削除します。

```
/path/to/happo-agent remove -g [案件コード[!サブグループ名]] -i [踏み台orNagiosサーバから見える自分自身のIPアドレス]
```


## Install

### Source-based install (Use upstart)

```bash
$ sudo yum install epel-release
$ sudo yum install nagios-plugins-all
$ go get -dv github.com/heartbeatsjp/happo-agent
$ cd $GOHOME/src/bin
$ openssl genrsa -aes128 -out happo-agent.key 2048
$ openssl req -new -key happo-agent.key -sha256 -out happo.csr
$ openssl x509 -in happo-agent.csr -days 3650 -req -signkey happo.key -sha256 -out happo.pub
$ touch metrics.yaml
$ chmod go-rwx happo-agent.key
$ sudo vim /etc/init/happo-agent.conf
$ sudo initctl reload-configuration
$ sudo initctl start happo-agent
```

happo-agent.conf

```
description "happo-agent"
author  "Your Name <USER@example.com>"

start on runlevel [2345]
stop on runlevel [016]

env LANG=C
env MARTINI_ENV=production
env APPNAME=happo-agent

exec /path/to/${APPNAME} daemon -A [Accept from IP/Subnet] -B ./${APPNAME}.pub -R ./${APPNAME}.key -M metrics.yaml >/dev/null 2>&1
respawn
```

Sensu メトリックプラグインは `/usr/local/bin` にインストールしてください。


### メトリック収集設定

`metrics.yaml`ファイルは、次の構造で記述します。1つのエージェントが、インストールしているサーバばかりでなく、リモートのメトリックを収集する設定を行うことも可能です。

```
metrics:
  - hostname: ホスト名
    plugins:
    - plugin_name: Sensuプラグイン実行形式ファイル名
      plugin_option: Sensuプラグインに与える引数
    - (以上同じ)
  - (以上同じ)
```


## API

- 待ち受けポート 6777 (Default)
- HTTPS, TLS 1.2以上対応 (従って古い`curl`では動かない)
- CipherSuitesは `TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256` を利用

### /

死活確認を行います。

- 入力形式
    - なし
- 返り値の形式
    - 文字列"OK"

```
$ wget -q --no-check-certificate -O - https://127.0.0.1:6777/
OK
```

### /proxy

エージェントが取得したメトリックデータを取得します。取得後、エージェントは溜め込んだメトリックデータを削除します。

- 入力形式
    - JSON
- 入力変数
    - proxy_hostport:
        - (配列) 踏み台サーバのIP:Port 多段可能
    - request_type: どのURIを実行するか
    - request_json: 監視対象サーバに届けるJSON
- 返り値の形式
    - JSON
- 返り値の変数
    - 実行したURLのレスポンスに従います

```
$ wget -q --no-check-certificate -O - https://192.0.2.1:6777/proxy --post-data='{"proxy_hostport": ["198.51.100.1:6777"], "request_type": "monitor", "request_json": "{\"apikey\": \"\", \"plugin_name\": \"check_procs\", \"plugin_option\": \"-w 100 -c 200\"}"}'
{"return_value":1,"message":"PROCS WARNING: 168 processes\n"}
```

例では次の通り呼び出しています: `wget host -> https://192.0.2.1:6777/proxy -> https://198.51.100.1:6777/monitor`.

### /inventory

インベントリ情報を取得します。

- 入力形式
    - JSON
- 入力変数
    - apikey: ""
    - command: 実行するコマンド
    - command_option: 実行するコマンドの引数
- 返り値の形式
    - JSON
- 返り値の変数
    - return_code: コマンド実行時の返り値
    - return_value: コマンド実行後の出力結果 (stdout, stderr)

```
$ wget -q --no-check-certificate -O - https://127.0.0.1:6777/inventory --post-data='{"apikey": "", "command": "uname", "command_option": "-a"}'
{"return_code":0,"return_value":"Linux saito-hb-vm101 2.6.32-573.3.1.el6.x86_64 #1 SMP Thu Aug 13 22:55:16 UTC 2015 x86_64 x86_64 x86_64 GNU/Linux\n"}
```

### /monitor

監視プラグインを実行します。

- 入力形式
    - JSON
- 入力変数
    - apikey: ""
    - plugin_name: 実行するコマンド
    - plugin_option: 実行するコマンドの引数
- 返り値の形式
    - JSON
- 返り値の変数
    - return_code: コマンド実行時の返り値
    - return_value: コマンド実行後の出力結果 (stdout, stderr)

```
$ wget -q --no-check-certificate -O - https://127.0.0.1:6777/monitor --post-data='{"apikey": "", "plugin_name": "check_procs", "plugin_option": "-w 100 -c 200"}'
{"return_value":1,"message":"PROCS WARNING: 168 processes\n"}
```

### /metric

エージェントが取得したメトリックデータを取得します。取得後、エージェントは溜め込んだメトリックデータを削除します。

- 入力形式
    - JSON
- 入力変数
    - apikey: ""
- 返り値の形式
    - JSON
- 返り値の変数
    - MetricData:
        - (配列)
            - hostname: ホスト名
            - timestamp: 取得秒 (UNIX)
            - metrics: 取得したメトリックの名前と値の組
    - Message: エージェントからのメッセージ (特にエラーがあれば掲載)

```
$ wget -q --no-check-certificate -O - https://127.0.0.1:6777/metric --post-data='{"apikey": ""}'
{"metric_data":[{"hostname":"saito-hb-vm101","timestamp":1444028730,"metrics":{"linux.context_switches.context_switches":32662,"linux.disk.elapsed.iotime_sda":52,"linux.disk.elapsed.iotime_weighted_sda":82,"linux.disk.rwtime.tsreading_sda":0,"linux.disk.rwtime.tswriting_sda":82,"linux.forks.forks":88,"linux.interrupts.interrupts":19642,"linux.ss.CLOSE-WAIT":0,"linux.ss.CLOSING":0,"linux.ss.ESTAB":9,"linux.ss.FIN-WAIT-1":0,"linux.ss.FIN-WAIT-2":0,"linux.ss.LAST-ACK":0,"linux.ss.LISTEN":31,"linux.ss.SYN-RECV":0,"linux.ss.SYN-SENT":0,"linux.ss.TIME-WAIT":7,"linux.ss.UNCONN":0,"linux.ss.UNKNOWN":0,"linux.swap.pswpin":0,"linux.swap.pswpout":0,"linux.users.users":1}},…(snip)…],"message":""}
```

### /metric/config/update

*TODO*

### /metric/status

Get collected metric status.

- Input format
    - None
- Input variables
    - None
- Return format
    - JSON
- Return variables
    - length: length of metric_data_buffer
    - capacity: capacity of metric_data_buffer
    - oldest_timestamp: oldest Timestamp(int64) in metric_data_buffer
    - newest_timestamp: newest Timestamp(int64) in metric_data_buffer

```
$ wget -q --no-check-certificate -O - https://127.0.0.1:6777/metric/status
{"capacity":4,"length":4,"newest_timestamp":1454654233,"oldest_timestamp":1454654173}
```

## Contribution

1. Fork ([http://github.com/heartbeatsjp/happo-agent/fork](http://github.com/heartbeatsjp/happo-agent/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create a new Pull Request


## Author

[Yuichiro Saito](https://github.com/koemu)
[Toshiaki Baba](https://github.com/netmarkjp)
