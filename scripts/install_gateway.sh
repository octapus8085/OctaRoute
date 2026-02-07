#!/usr/bin/env bash
set -euo pipefail

PREFIX=${PREFIX:-/usr/local}
BIN_DIR="$PREFIX/bin"
SERVICE_NAME=octaroute-gatewayd

mkdir -p "$BIN_DIR"

go build -o "$BIN_DIR/$SERVICE_NAME" ./cmd/octaroute-gatewayd

if command -v systemctl >/dev/null 2>&1; then
  echo "Binary installed to $BIN_DIR/$SERVICE_NAME"
  echo "To enable systemd: sudo cp systemd/$SERVICE_NAME.service /etc/systemd/system/ && sudo systemctl daemon-reload"
fi
