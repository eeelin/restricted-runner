# runner utilities

This directory contains caller-side helper scripts intended for use inside a GitHub Actions runner image.

## Files

- `rr-exec`

## Primary helper

`rr-exec` is the main supported helper in the first version.
It builds a structured request and sends it over SSH stdin to a remote host running `restricted-runner` behind a forced-command entrypoint.

`--target` means the logical policy target, not a path.
Use `--dry-run` when you want remote preflight without real execution and your remote SSH entrypoint is wired to convert that into `dispatch --dry-run`.

## Example

```bash
rr-exec \
  --host deploy@my-host \
  --caller github-actions-homecloud \
  --target server \
  --script homecloud/site/apply \
  --arg sites/homes/ruyi/hass \
  --env TARGET=server \
  --env ACTOR=github-actions
```
