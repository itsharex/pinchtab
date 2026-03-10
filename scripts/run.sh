#!/bin/bash
# Run an existing PinchTab binary from the repo root.
set -euo pipefail

cd "$(dirname "$0")/.."

if [ ! -x ./pinchtab ]; then
  echo "pinchtab binary not found at ./pinchtab"
  echo "Build it first with: go build -o pinchtab ./cmd/pinchtab"
  exit 1
fi

exec ./pinchtab "$@"
