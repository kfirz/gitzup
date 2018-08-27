#!/usr/bin/env bash

INOTIFY=$(which inotifywait 2>/dev/null)
[[ -z "${INOTIFY}" ]] && echo "Install inotify-tools first." >&2 && exit 1

function build {
    RESOURCES=$(find ./cmd/resources -name '*.go' -type f | sed 's|^./cmd/resources/||' | sed 's|.go$||')
    for r in $(find ./cmd/resources -name '*.go' -type f | sed 's|^./cmd/resources/||' | sed 's|.go$||'); do
        docker build --build-arg target=${r} --tag gitzup/${r}:dev --file ./build/Dockerfile .
    done
}

function stop {
    pkill -P $$
}

trap stop EXIT

build

while true; do
    EVENT=$(inotifywait -e create,modify,delete -r -q ./Makefile ./api ./cmd/resources ./internal)
    echo ""
    echo "====[ $(date) ]======================================================"
    echo "Change detected: ${EVENT}"
    echo "====================================================================="
    echo ""
    build
done
