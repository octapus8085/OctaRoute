#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

check_binary() {
  local bin=$1
  if [ -x "$bin" ]; then
    echo "ok: $bin"
  else
    echo "missing: $bin"
  fi
}

check_binary "$ROOT_DIR/cmd/octaroute-exitd"
check_binary "$ROOT_DIR/cmd/octaroute-gatewayd"
check_binary "$ROOT_DIR/cmd/octaroute-controller"

if [ -d "$ROOT_DIR/web" ]; then
  echo "ok: web directory present"
else
  echo "missing: web directory"
fi

if [ -d "$ROOT_DIR/systemd" ]; then
  echo "ok: systemd unit files present"
else
  echo "missing: systemd unit files"
fi
