# How to use docker in local development environment

- Requirements
    - Docker
- Local Requirements
    - yq ( use at Setup (to specify docker image name) )
    - jq ( yq requires jq )

Local Workdir is project root ( same as main.go )

## Setup

```
docker run --rm -it -d --security-opt=seccomp:unconfined -v $(pwd):/mnt -p 6777:6777 -p 2345:2345 --name=happo-agent-dev $(cat wercker.yml | yq ".box" -r) /mnt/contrib/development/docker_run.sh
```

=> may took few minutes. if you want to know progress, use `docker logs -f happo-agent-dev`

## Run all test

```
docker exec -it happo-agent-dev /mnt/contrib/development/docker_exec_test.sh
```

## Run specified test(ex: in case collect package)

```
docker exec -it -w /go/src/github.com/heartbeatsjp/happo-agent happo-agent-dev go test ./collect/...
```

## Debug with gdb-like shell (ex: in case `daemon` subcommand)

```
docker exec -it -w /go/src/github.com/heartbeatsjp/happo-agent happo-agent-dev dlv debug -- daemon -A 0.0.0.0/0 -B /go/happo-agent.pub -R /go/happo-agent.key -M /go/metrics.yaml
```


## Debug with remote (ex: in case `daemon` subcommand)

```
docker exec -it -w /go/src/github.com/heartbeatsjp/happo-agent happo-agent-dev dlv debug --headless --listen=:2345 --log -- daemon -A 0.0.0.0/0 -B /go/happo-agent.pub -R /go/happo-agent.key -M /go/metrics.yaml
```

## How to see happo-agent.log

```
docker exec -it -w /go/src/github.com/heartbeatsjp/happo-agent happo-agent-dev tail -F happo-agent.log
```

