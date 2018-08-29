#!/usr/bin/env bash

TAG=${1}
[[ -z "${TAG}" ]] && echo "usage: $0 <tag>" >&2 && exit 1

for i in $(find ./cmd/resources -name '*.go' -type f | sed 's|^./cmd/resources/||' | sed 's|.go$||'); do
    docker build --build-arg resource=${i} --tag gitzup/${i}:${TAG} --file $(dirname $0)/../Dockerfile.resource .
    docker push gitzup/${i}:${TAG}
    docker tag gitzup/${i}:${TAG} gitzup/${i}:latest
    docker push gitzup/${i}:latest
done
