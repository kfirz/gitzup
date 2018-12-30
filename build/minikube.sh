#!/usr/bin/env bash

set -eu -o pipefail

# Read current local version, defaulting to 1
VERSION="1"
if [[ -f .version ]]; then
    VERSION="$(cat .version)"
fi

# Increment and save back
VERSION="$((VERSION + 1))"
echo -n "${VERSION}" > .version
IMAGE="gitzup-manager:${VERSION}"

# Switch context to Minikube
eval $(minikube docker-env)

# Build image
docker build . -t ${IMAGE} -f ./build/Dockerfile

# Create a patch
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

# Deploy
./build/create_manifest.sh local | kubectl apply -f -
