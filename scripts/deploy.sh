#!/usr/bin/env bash
set -euo pipefail

# This is a minimal example deploy script.
# Adjust paths/host/service name to your environment.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

TARGET_HOST="${TARGET_HOST:-}"
TARGET_DIR="${TARGET_DIR:-/opt/kxl_backend_go}"
SERVICE_NAME="${SERVICE_NAME:-kxl-backend-go}"

if [[ -z "$TARGET_HOST" ]]; then
  echo "TARGET_HOST is required (e.g. user@server)"
  exit 1
fi

bash "$ROOT_DIR/scripts/build.sh"

echo "Deploying to $TARGET_HOST:$TARGET_DIR"
ssh "$TARGET_HOST" "mkdir -p '$TARGET_DIR'"
rsync -av --delete \
  "$ROOT_DIR/dist/" \
  "$ROOT_DIR/config/" \
  "$ROOT_DIR/templates/" \
  "$ROOT_DIR/static/" \
  "$TARGET_HOST:$TARGET_DIR/"

echo "Restarting systemd service: $SERVICE_NAME"
ssh "$TARGET_HOST" "sudo systemctl restart '$SERVICE_NAME'"

