#!/usr/bin/env bash

$(dirname $0)/watch-and-rerun-commands.sh api-server GREEN &
$(dirname $0)/watch-and-rerun-commands.sh buildagent CYAN &
$(dirname $0)/watch-and-rerun-commands.sh webhooks-server YELLOW &
$(dirname $0)/watch-and-rebuild-resources.sh BLUE &

function stop {
    pkill -P $$
}

trap stop EXIT

wait
