#!/bin/sh
set -eu

PREFIX=${PREFIX:-/usr/local}
CONFIG_DIR=${CONFIG_DIR:-/etc/restricted-runner}
ROOT_DIR=${ROOT_DIR:-/opt/restricted-runner/root}
BIN_DIR="$PREFIX/bin"

rm -f "$BIN_DIR/restricted-runner-ssh-entrypoint"
rm -f "$ROOT_DIR/homecloud/site/validate"
rm -f "$ROOT_DIR/homecloud/site/apply"
rmdir "$ROOT_DIR/homecloud/site" 2>/dev/null || true
rmdir "$ROOT_DIR/homecloud" 2>/dev/null || true
rmdir "$ROOT_DIR" 2>/dev/null || true
rm -f "$CONFIG_DIR/config.yaml"
rmdir "$CONFIG_DIR" 2>/dev/null || true

cat <<EOF
removed demo files:
- $BIN_DIR/restricted-runner-ssh-entrypoint
- $ROOT_DIR/homecloud/site/validate
- $ROOT_DIR/homecloud/site/apply
- $CONFIG_DIR/config.yaml

note:
- this script does not edit authorized_keys
- remove any forced-command key entries manually if needed
EOF
