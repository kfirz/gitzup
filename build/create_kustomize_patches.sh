#!/usr/bin/env bash

IMAGE=${1}
[[ -z "${IMAGE}" ]] && echo "usage: $0 <image>" && exit 1

set -eu -o pipefail

cat > ./config/default/manager_image_patch.yaml <<EOF
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: controller-manager
  namespace: gitzup
spec:
  template:
    spec:
      containers:
      - name: manager
        image: ${IMAGE}
EOF
