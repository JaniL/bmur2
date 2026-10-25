#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

rm -rf artifacts
mkdir -p artifacts

source_files=(go.mod go.sum config.go handler.go bot.go lastfm.go format.go)

if command -v go >/dev/null 2>&1; then
  go test ./...
  go vet ./...
  go mod tidy
  go mod vendor
  source_files+=(vendor)
fi

zip -r artifacts/function.zip "${source_files[@]}"

echo "created artifacts/function.zip"
