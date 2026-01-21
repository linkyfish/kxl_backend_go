#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

OUT_DIR="${OUT_DIR:-dist}"
BINARY="${BINARY:-kxl-api}"

GOOS="${GOOS:-linux}"
GOARCH="${GOARCH:-amd64}"
CGO_ENABLED="${CGO_ENABLED:-0}"

mkdir -p "$OUT_DIR"

echo "Building $BINARY ($GOOS/$GOARCH) -> $OUT_DIR/$BINARY"
CGO_ENABLED="$CGO_ENABLED" GOOS="$GOOS" GOARCH="$GOARCH" go build -o "$OUT_DIR/$BINARY" ./cmd/api

