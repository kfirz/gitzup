#!/usr/bin/env bash

./build/scripts/run.sh api-server GREEN &
./build/scripts/run.sh buildagent CYAN &
./build/scripts/run.sh webhooks-server YELLOW &
./build/scripts/resources.sh BLUE &

function stop {
    pkill -P $$
}

trap stop EXIT

wait
