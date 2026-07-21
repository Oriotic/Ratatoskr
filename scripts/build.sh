#!/usr/bin/env bash
# Builds the ratatoskr binary into ./bin/ratatoskr
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p bin
go build -o bin/ratatoskr .
echo "Built bin/ratatoskr"
