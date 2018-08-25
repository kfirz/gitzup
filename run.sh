#!/usr/bin/env bash

docker-compose --file=./deployments/docker-compose.yml \
               --project-name=gitzup \
               --project-directory=$(pwd) \
               up \
               --quiet-pull \
               --build \
               --abort-on-container-exit \
               --remove-orphans
