# restricted-runner

A restricted execution helper for self-hosted runner environments.

This project is implemented in Go and follows a CGI-like execution model over a restricted transport boundary.
Design documents live under `docs/`.

## Documents

- `docs/design.md`
- `docs/protocol.md`
- `docs/config.md`

## Example files

- `examples/config.yaml`
- `examples/root/homecloud/site/validate`
- `examples/root/homecloud/site/apply`

## Example usage

Validate only:

```bash
cat <<'EOF' | restricted-runner validate \
  --config ./examples/config.yaml \
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
  --config ./examples/config.yaml \
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
