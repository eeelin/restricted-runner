# restricted-runner

A restricted execution helper for self-hosted runner environments.

This project is implemented in Go and follows a CGI-like execution model over a restricted transport boundary.
Design documents live under `docs/`.

## Documents

- `docs/design.md`
- `docs/protocol.md`
- `docs/config.md`
- `docs/ssh.md`

## CI and release

- pull requests and pushes to `main` run `gofmt`, `go test ./...`, and `go build ./cmd/restricted-runner`
- tags matching `v*` build a Linux x86-64 release binary and publish release artifacts

## Additional documents

- `docs/runner-image.md`
- `docs/ssh.md`

## Example files

- `examples/config.yaml`
- `examples/root/homecloud/site/validate`
- `examples/root/homecloud/site/apply`
- `examples/ssh/authorized_keys.example`
- `examples/ssh/restricted-runner-ssh-entrypoint`

## Demo install scripts

- `scripts/install-demo.sh`
- `scripts/uninstall-demo.sh`

## Example usage

The sample config is a template. Before using it, replace the `root_path` value with an absolute path to `examples/root` on your machine.

For example:

```bash
ROOT=$(pwd)/examples/root
sed "s#ROOT_PATH_PLACEHOLDER#$ROOT#" examples/config.yaml > /tmp/restricted-runner.config.yaml
```

Validate only:

```bash
cat <<'EOF' | restricted-runner validate \
  --config /tmp/restricted-runner.config.yaml \
  --caller github-actions-homecloud \
  --target server
{
  "version": "v1",
  "request_id": "req-123",
  "script": "homecloud/site/validate",
  "argv": ["sites/homes/ruyi/hass"],
  "env": {
    "TARGET": "server",
    "ACTOR": "eeelin"
  },
  "stdin": "hello from stdin"
}
EOF
```

Preflight dispatch:

```bash
cat <<'EOF' | restricted-runner dispatch --dry-run \
  --config /tmp/restricted-runner.config.yaml \
  --caller github-actions-homecloud \
  --target server
{
  "version": "v1",
  "request_id": "req-124",
  "script": "homecloud/site/apply",
  "argv": ["sites/homes/ruyi/hass"],
  "env": {
    "TARGET": "server",
    "ACTOR": "eeelin"
  }
}
EOF
```

## Release output

Current release automation builds:

- `restricted-runner-linux-amd64`
- `restricted-runner-linux-amd64.tar.gz`
- `sha256sums.txt`

## Runner image and SSH utilities

- `docker/runner/Dockerfile`
- `scripts/runner/rr-exec`
- `scripts/runner/README.md`
