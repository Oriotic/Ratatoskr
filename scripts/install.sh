#!/usr/bin/env bash
# Builds ratatoskr and installs it to /usr/local/bin (or $PREFIX/bin).
set -euo pipefail
cd "$(dirname "$0")/.."
PREFIX="${PREFIX:-/usr/local}"
go build -o /tmp/ratatoskr-build .
sudo install -m 0755 /tmp/ratatoskr-build "$PREFIX/bin/ratatoskr"
rm -f /tmp/ratatoskr-build
echo "Installed to $PREFIX/bin/ratatoskr"
