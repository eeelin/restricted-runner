# runner utilities

This directory contains caller-side helper scripts intended for use inside a GitHub Actions runner image.

## Files

- `rr-dispatch-ssh`
- `rr-validate-ssh`

## Primary helper

`rr-dispatch-ssh` is the main supported helper in the first version.
It builds a structured request and sends it over SSH stdin to a remote host running `restricted-runner` behind a forced-command entrypoint.

## Example

```bash
rr-dispatch-ssh \
  --host deploy@my-host \
  --caller github-actions-homecloud \
  --target server \
  --script homecloud/site/apply \
  --arg sites/homes/ruyi/hass \
  --env TARGET=server \
  --env ACTOR=github-actions
```
