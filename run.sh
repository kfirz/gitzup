#!/usr/bin/env bash

./build/scripts/run.sh api-server &
./build/scripts/run.sh buildagent &
./build/scripts/run.sh webhooks-server &
./build/scripts/resources.sh &

function stop {
    pkill -P $$
}

trap stop EXIT

wait
