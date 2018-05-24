#!/bin/bash

export PATH=$PATH:/usr/local/bin

if [[ ! -f /root/setup ]]; then
    echo "ERROR: setup not yet finished. retry after few seconds."
    exit 1
fi

cd /go/src/github.com/heartbeatsjp/happo-agent/
echo "### rsync"
rsync -av --delete --exclude="vendor" --exclude=".wercker" /mnt/ /go/src/github.com/heartbeatsjp/happo-agent/
echo

echo "### wercker steps"

echo "cd /go/src/github.com/heartbeatsjp/happo-agent/; $(cat /mnt/wercker.yml | yq '.build.steps[:-1][].script.code' -r | grep -v '^null$')" | bash -xe
