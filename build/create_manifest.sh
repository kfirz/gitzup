#!/usr/bin/env bash

BASE=${1}
if [[ -z "${BASE}" ]]; then
    BASE="default"
fi

set -eu -o pipefail

# Concatenate the CRD assets
for crdFile in $(ls ./config/crds/*.yaml); do
    cat ${crdFile}
    echo ""
    echo "---"
done

# Append the Gitzup deployment objects
kustomize build ./config/${BASE}
