#!/usr/bin/env bash

set -eu -o pipefail

source $(cd $(dirname $0); pwd)/manifest_crds.sh
source $(cd $(dirname $0); pwd)/manifest_manager.sh $@
