#!/usr/bin/env bash

INOTIFY=$(which inotifywait 2>/dev/null)
[[ -z "${INOTIFY}" ]] && echo "Install inotify-tools first." >&2 && exit 1

TARGET="${1}"
[[ -z "${TARGET}" ]] && echo "usage: $0 <target>" >&2 && exit 1

function start {
    make ${TARGET}
    if [[ $? == 0 ]]; then
        echo "Starting ${TARGET}..." >&2
        GOOGLE_APPLICATION_CREDENTIALS=$(ls *-sa-*.local.json|head -1) ./${TARGET} &
    else
        echo "Failed building ${TARGET}" >&2
    fi
}

function stop {
    pkill -P $$
}

function restart {
    stop
    start
}

trap stop EXIT

start

while true; do
    EVENT=$(inotifywait -e create,modify,delete -r -q ./Makefile ./api ./cmd/${TARGET} ./internal ./web)
    [[ $? != 0 ]] && exit 1
    echo "" >&2
    echo "=====================================================================" >&2
    echo "Change detected: ${EVENT}" >&2
    restart
done
