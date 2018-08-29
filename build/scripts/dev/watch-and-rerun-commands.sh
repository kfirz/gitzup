#!/usr/bin/env bash

INOTIFY=$(which inotifywait 2>/dev/null)
TARGET="${1}"
COLOR="${2:-NORMAL}"
JSON_KEY=$(ls *-sa-*.local.json|head -1)
PREFIX=$(printf "%15s\n" "${TARGET}")
[[ -z "${INOTIFY}" ]] && echo "Install inotify-tools first." >&2 && exit 1
[[ -z "${TARGET}" ]] && echo "usage: $0 <target>" >&2 && exit 1

function watchAndInterrupt {
    while true; do
        EVENT=$(inotifywait -e create,modify,delete -t 1 -r -q ./Makefile ./api ./cmd/${TARGET} ./internal ./web)
        RC=$?
        if [[ ${RC} == 0 ]]; then
            pkill --parent $$ --exact ${TARGET}
        elif [[ ${RC} != 2 ]]; then
            echo "File watcher failed! (${RC})" 2>&1 | $(dirname $0)/../util/colorize.sh ${COLOR} "  ${PREFIX}: "
            exit 1
        fi
    done
}

function stop {
    pkill -P $$
}

trap stop EXIT

watchAndInterrupt &

while true; do
    make ${TARGET} 2>&1 | $(dirname $0)/../util/colorize.sh ${COLOR} "  ${PREFIX}: "
    if [[ $? == 0 ]]; then
        GOOGLE_APPLICATION_CREDENTIALS=${JSON_KEY} ./${TARGET} 2>&1 | $(dirname $0)/../util/colorize.sh ${COLOR} "  ${PREFIX}: "
    else
        echo "Failed building ${TARGET}" 2>&1 | $(dirname $0)/../util/colorize.sh ${COLOR} "  ${PREFIX}: "
        exit 1
    fi
done
