#!/usr/bin/env bash

set -eu -o pipefail

kustomize build ./config/${1:-default}
