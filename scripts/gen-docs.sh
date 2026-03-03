#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="$REPO_ROOT/docs/api"

mkdir -p "$OUT_DIR"

cd "$REPO_ROOT"

gomarkdoc \
  --output "$OUT_DIR/API.md" \
  ./...

echo "Generated $OUT_DIR/API.md"
