#!/usr/bin/env bash

CMD=${1}
[[ -z "${CMD}" ]] && echo "usage: $0 <cmd> <tag>" >&2 && exit 1
TAG=${2}
[[ -z "${TAG}" ]] && echo "usage: $0 <cmd> <tag>" >&2 && exit 1

docker build --tag gitzup/${CMD}:${TAG} --file $(dirname $0)/../Dockerfile.${CMD} .
docker push gitzup/${CMD}:${TAG}
docker tag gitzup/${CMD}:${TAG} gitzup/${CMD}:latest
docker push gitzup/${CMD}:latest
