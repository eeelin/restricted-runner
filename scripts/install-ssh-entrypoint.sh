#!/bin/sh
set -eu

PREFIX=${PREFIX:-/usr/local}
CONFIG_PATH=${CONFIG_PATH:-/etc/restricted-runner/config.yaml}
BIN_DIR="$PREFIX/bin"
REPO_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)

install -d "$BIN_DIR"
install -m 0755 "$REPO_DIR/examples/ssh/restricted-runner-ssh-entrypoint" "$BIN_DIR/restricted-runner-ssh-entrypoint"

cat <<EOF
installed:
- $BIN_DIR/restricted-runner-ssh-entrypoint

required sshd_config support:
- AcceptEnv RR_CALLER RR_TARGET RR_PREFLIGHT

expected runtime env:
- RESTRICTED_RUNNER_BIN=/usr/local/bin/restricted-runner (optional override)
- RR_CONFIG_PATH=${CONFIG_PATH}
- RR_CALLER=<caller-id>
- RR_TARGET=<logical-target>
- RR_PREFLIGHT=1 (optional, for preflight mode)
EOF
