# Dockerfile for happo-agent development environment (mainly local development)
#
# Usage:
#     docker build -t happo-agent-dev .
#     docker run --security-opt=seccomp:unconfined --rm -it -p 6777:6777 -v $(pwd):/mnt -w /go/src/github.com/heartbeatsjp/happo-agent --name happo-agent-dev happo-agent-dev /bin/bash
#         rsync -av --exclude="vendor/" --exclude=".wercker/" /mnt/ /go/src/github.com/heartbeatsjp/happo-agent/ && dep ensure
FROM netmarkjp/golang-build:1.9.2

RUN curl -L -o /tmp/mackerel-agent_latest.all.deb https://mackerel.io/file/agent/deb/mackerel-agent_latest.all.deb \
    && dpkg -i /tmp/mackerel-agent_latest.all.deb \
    && rm -f /tmp/mackerel-agent_latest.all.deb
RUN echo "deb http://apt.mackerel.io/v2/ mackerel contrib" > /etc/apt/sources.list.d/mackerel.list
RUN curl -LfsS https://mackerel.io/file/cert/GPG-KEY-mackerel-v2 | apt-key add -
RUN apt-get update -qq \
    && apt-get install -y \
        vim \
        less \
        rsync \
        git \
        mackerel-agent-plugins \
        nagios-plugins \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*
RUN mkdir -p /go/src/github.com/heartbeatsjp/happo-agent
ENV GOPATH /go

RUN mkdir /etc/happo-agent

RUN openssl genrsa -out /etc/happo-agent/happo-agent.key 2048 
RUN yes "" | openssl req -new -key /etc/happo-agent/happo-agent.key -sha256 -out /etc/happo-agent/happo-agent.csr
RUN openssl x509 -in /etc/happo-agent/happo-agent.csr -days 3650 -req -signkey /etc/happo-agent/happo-agent.key -sha256 -out /etc/happo-agent/happo-agent.pub
ENV HAPPO_AGENT_PUBLIC_KEY  /etc/happo-agent/happo-agent.pub
ENV HAPPO_AGENT_PRIVATE_KEY /etc/happo-agent/happo-agent.key

RUN touch /etc/happo-agent/metrics.yaml
ENV HAPPO_AGENT_METRIC_CONFIG /etc/happo-agent/metrics.yaml

ENV HAPPO_AGENT_ALLOWED_HOSTS 127.0.0.1,172.17.0.1
ENV HAPPO_AGENT_LOGFILE /var/log/happo-agent.log
ENV HAPPO_AGENT_DBFILE /var/lib/happo-agent.db
ENV HAPPO_AGENT_SENSU_PLUGIN_PATHS /usr/local/hb-agent/bin,/usr/local/bin,/usr/bin

EXPOSE 6777
