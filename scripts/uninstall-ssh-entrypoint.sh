#!/bin/sh
set -eu

PREFIX=${PREFIX:-/usr/local}
BIN_DIR="$PREFIX/bin"

rm -f "$BIN_DIR/restricted-runner-ssh-entrypoint"

echo "removed: $BIN_DIR/restricted-runner-ssh-entrypoint"
