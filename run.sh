#!/usr/bin/env bash

function start {
    docker-compose --file=./deployments/docker-compose.yml \
                   --project-name=gitzup \
                   --project-directory=$(pwd) \
                   up \
                   --detach \
                   --quiet-pull \
                   --build \
                   --remove-orphans
    [[ "$?" != "0" ]] && echo "Failed starting compose" && exit 1
}

function stop {
    docker-compose --file=./deployments/docker-compose.yml \
                   --project-name=gitzup \
                   --project-directory=$(pwd) \
                   down
}

function restart {
    stop
    start
}

trap stop EXIT

INOTIFY=$(which inotifywait 2>/dev/null)
[[ -z "${INOTIFY}" ]] && echo "Install inotify-tools first." >&2 && exit 1

start

echo "Now watching $(pwd)"
while true; do
    EVENT=$(inotifywait -e create,modify,delete -r -q ./api ./build ./cmd ./deployments ./internal ./web ./Gopkg.*)
    echo ""
    echo "====[ $(date) ]======================================================"
    echo "Change detected: ${EVENT}"
    echo "====================================================================="
    echo ""
    restart
done
