#!/bin/sh
set -e

if [ -z "$INSTANCE_ID" ] && [ -S /var/run/docker.sock ]; then
  CID=$(hostname)
  NUM=$(curl -sf --unix-socket /var/run/docker.sock "http://localhost/containers/${CID}/json" \
    | sed -n 's/.*"com.docker.compose.container-number":"\([^"]*\)".*/\1/p' | head -1)
  if [ -n "$NUM" ]; then
    export INSTANCE_ID="api/backend_${NUM}"
  fi
fi

exec ./keeneye_app
