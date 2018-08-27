#!/usr/bin/env bash

COLOR="${1:-NORMAL}"
INOTIFY=$(which inotifywait 2>/dev/null)
PREFIX=$(printf "%15s\n" "resources")
[[ -z "${INOTIFY}" ]] && echo "Install inotify-tools first." >&2 && exit 1

function build {
    for r in $(find ./cmd/resources -name '*.go' -type f | sed 's|^./cmd/resources/||' | sed 's|.go$||'); do
        docker build --quiet \
                     --build-arg target=${r} \
                     --tag gitzup/${r}:dev \
                     --file \
                     ./build/Dockerfile . 2>&1 \
             | ./build/scripts/colorize.sh ${COLOR} "  ${PREFIX}: "
    done
}

function stop {
    pkill -P $$
}

trap stop EXIT

build

while true; do
    EVENT=$(inotifywait -e create,modify,delete -r -q ./Makefile ./api ./cmd/resources ./internal)
    build
done
