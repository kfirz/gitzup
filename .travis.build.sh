#!/usr/bin/env bash

set -exu -o pipefail

if [[ -n "${TRAVIS_TAG}" ]]; then
    # TODO: only run "latest" if this tag is the latest GitHub release tag (use "docker" otherwise since we're building an older one)
    PUSH=true TAG=${TRAVIS_TAG} make latest

elif [[ -n "${TRAVIS_PULL_REQUEST_SHA}" ]]; then
    PUSH=true REPO=gcr.io/gitzup TAG=${TRAVIS_PULL_REQUEST_SHA} make docker

else
    PUSH=true REPO=gcr.io/gitzup TAG=${TRAVIS_COMMIT} make docker
fi
