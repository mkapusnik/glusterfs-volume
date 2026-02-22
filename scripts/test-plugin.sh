#!/usr/bin/env bash
set -euo pipefail

VOLUME_NAME=${VOLUME_NAME:-configs}
PLUGIN_NAME=${PLUGIN_NAME:-glusterfs-volume}

SERVERS=${SERVERS:-gfs.lan}
if [[ -z "${SERVERS}" ]]; then
  echo "SERVERS must be set to existing GlusterFS servers." >&2
  exit 1
fi

if docker plugin inspect "${PLUGIN_NAME}" >/dev/null 2>&1; then
  docker plugin disable --force "${PLUGIN_NAME}"
  docker plugin rm --force "${PLUGIN_NAME}"
fi

make build PLUGIN="${PLUGIN_NAME}"

docker plugin enable "${PLUGIN_NAME}"

docker volume create \
  --driver "${PLUGIN_NAME}" \
  --opt server="${SERVERS}" \
  --opt volume="${VOLUME_NAME}" \
  "${VOLUME_NAME}"

docker run --rm \
  -v "${VOLUME_NAME}:/data" \
  alpine:3.19 sh -c "ls -a /data && test -n \"$(ls -A /data)\""

docker volume rm "${VOLUME_NAME}"
if docker plugin inspect "${PLUGIN_NAME}" >/dev/null 2>&1; then
  docker plugin disable --force "${PLUGIN_NAME}"
  docker plugin rm --force "${PLUGIN_NAME}"
fi
