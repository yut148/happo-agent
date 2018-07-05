#!/bin/bash

cd /go
mkdir -p src/github.com/heartbeatsjp/happo-agent || :
apt-get update
apt-get -y install rsync python-pip jq nagios-plugins
pip install yq

curl -L -o /tmp/mackerel-plugins.deb https://github.com/mackerelio/mackerel-agent-plugins/releases/download/v0.49.0/mackerel-agent-plugins_0.49.0-1.v2_amd64.deb
dpkg -i /tmp/mackerel-plugins.deb
ln -s /usr/bin/mackerel-plugin* /usr/local/bin/.

install -d -m 755 /root/.ssh
cat << EOT | tee /root/.ssh/config
Host *
    UserKnownHostsFile /dev/null
    StrictHostKeyChecking no
EOT

openssl genrsa -out happo-agent.key 2048
yes "" | openssl req -new -key happo-agent.key -sha256 -out happo-agent.csr
openssl x509 -in happo-agent.csr -days 3650 -req -signkey happo-agent.key -sha256 -out happo-agent.pub
touch metrics.yaml

touch /root/setup

echo "all things done. start..."
exec /bin/bash
