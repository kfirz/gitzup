#!/usr/bin/env bash

set -eu -o pipefail

echo "---> CHECKING MINIKUBE IS UP & RUNNING"
minikube status
echo

echo -n "---> TARGET IMAGE NAME: "
VERSION="1"
if [[ -f .version ]]; then
    VERSION="$(cat .version)"
fi
VERSION="$((VERSION + 1))"
echo -n "${VERSION}" > .version
IMAGE="gitzup-manager:${VERSION}"
echo "${IMAGE}"

echo "---> SWITCHING DOCKER HOST TO MINIKUBE"
eval $(minikube docker-env)

echo "---> BUILDING DOCKER IMAGE"
docker build . -t ${IMAGE} -f ./build/Dockerfile

echo "---> SAVING IMAGE IN Kustomize PATCH"
source $(cd $(dirname $0); pwd)/create_kustomize_patches.sh ${IMAGE}

echo "---> DEPLOYING TO CLUSTER"
./build/manifest.sh local | kubectl apply -f -
