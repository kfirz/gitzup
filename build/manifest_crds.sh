#!/usr/bin/env bash

set -eu -o pipefail

# Concatenate the CRD assets
for crdFile in $(ls ./config/crds/*.yaml); do
    cat ${crdFile}
    echo ""
    echo "---"
done
