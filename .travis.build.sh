#!/usr/bin/env bash

set -exu -o pipefail

if [[ -n "${TRAVIS_TAG}" ]]; then
    # TODO: only run "latest" if this tag is the latest GitHub release tag (use "docker" otherwise since we're building an older one)
    TAG=${TRAVIS_TAG} make push-docker push-latest

elif [[ -n "${TRAVIS_PULL_REQUEST_SHA}" ]]; then
    REPO=gcr.io/gitzup TAG=${TRAVIS_PULL_REQUEST_SHA} make push-docker

else
    REPO=gcr.io/gitzup TAG=${TRAVIS_COMMIT} make push-docker
fi
