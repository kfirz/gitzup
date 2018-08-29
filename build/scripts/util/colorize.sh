#!/usr/bin/env bash

GREEN="\033[0;32m"
CYAN="\033[0;36m"
GRAY="\033[0;37m"
BLUE="\033[0;34m"
YELLOW="\033[0;33m"
NORMAL="\033[m"

color=\$${1:-NORMAL}
prefix=${2}

function stop {
    pkill -P $$
}

trap stop EXIT

cat | while read LINE; do
    echo -e "${prefix}$(eval echo ${color})${LINE}${NORMAL}"
done
