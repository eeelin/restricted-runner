#!/bin/sh
set -eu

PREFIX=${PREFIX:-/usr/local}
CONFIG_DIR=${CONFIG_DIR:-/etc/restricted-runner}
ROOT_DIR=${ROOT_DIR:-/opt/restricted-runner/root}
BIN_DIR="$PREFIX/bin"
EXAMPLE_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)

mkdir -p "$BIN_DIR" "$CONFIG_DIR" "$ROOT_DIR/homecloud/site"

if [ ! -f "$BIN_DIR/restricted-runner" ]; then
  echo "warning: $BIN_DIR/restricted-runner does not exist yet"
  echo "build or install the binary first, for example:"
  echo "  go build -o $BIN_DIR/restricted-runner ./cmd/restricted-runner"
fi

install -m 0755 "$EXAMPLE_DIR/examples/ssh/restricted-runner-ssh-entrypoint" "$BIN_DIR/restricted-runner-ssh-entrypoint"
install -m 0755 "$EXAMPLE_DIR/examples/root/homecloud/site/validate" "$ROOT_DIR/homecloud/site/validate"
install -m 0755 "$EXAMPLE_DIR/examples/root/homecloud/site/apply" "$ROOT_DIR/homecloud/site/apply"

sed "s#ROOT_PATH_PLACEHOLDER#$ROOT_DIR#" "$EXAMPLE_DIR/examples/config.yaml" > "$CONFIG_DIR/config.yaml"

cat <<EOF
installed demo files:
- $BIN_DIR/restricted-runner-ssh-entrypoint
- $ROOT_DIR/homecloud/site/validate
- $ROOT_DIR/homecloud/site/apply
- $CONFIG_DIR/config.yaml

next steps:
1. install or build the restricted-runner binary into $BIN_DIR/restricted-runner
2. add examples/ssh/authorized_keys.example content to the restricted user's authorized_keys
3. adjust RR_CALLER / RR_TARGET / RR_CONFIG_PATH as needed
EOF
