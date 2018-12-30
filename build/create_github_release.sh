#!/usr/bin/env bash

[[ -z "${TAG_NAME}" ]] && echo "TAG_NAME is empty" && exit 1

set -eu -o pipefail

# Create the GitHub release and extract its upload URL
cat > curl_headers <<EOF
Authorization: token $(cat ./github-access-token)
Content-Type: application/yaml
EOF
curl -sSL -H @curl_headers "https://api.github.com/repos/kfirz/gitzup/releases/tags/${TAG_NAME}" > release.json

# Upload the resulting Gitzup file
UPLOAD_URL="$(cat ./release.json | jq -r '.upload_url | split("{")[0]')"
curl -sSL -H @curl_headers -X POST --data-binary @gitzup.yaml "${UPLOAD_URL}?name=gitzup.yaml"
