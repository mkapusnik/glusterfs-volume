#!/usr/bin/env bash
set -euo pipefail

PLUGIN_NAME=${PLUGIN_NAME:-glusterfs-volume}
REGISTRY=${REGISTRY:-}
CONTEXT=${DOCKER_CONTEXT:-default}

FULL_NAME="${REGISTRY}${PLUGIN_NAME}"

echo "Building and pushing ${FULL_NAME}"
make build PLUGIN="${PLUGIN_NAME}" REGISTRY="${REGISTRY}" CONTEXT="${CONTEXT}"
docker --context "${CONTEXT}" plugin push "${FULL_NAME}"
