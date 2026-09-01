# runner utilities

This directory contains caller-side helper scripts intended for use inside a GitHub Actions runner image.

## Files

- `rr-exec`
- `start-runner`
- `start-runner_test.sh`

## Primary helper

`rr-exec` is the main supported helper in the first version.
It builds a structured request and sends it over SSH stdin to a remote host running `restricted-runner` behind a forced-command entrypoint.

`--target` means the logical policy target, not a path.
Use `--preflight` when you want remote dispatch preflight without real execution.

## Runner startup

`start-runner` registers the GitHub Actions runner when no saved state exists,
copies the generated identity files into `RUNNER_STATE_DIR`, and starts the
runner. On later launches it restores those files and skips registration, so an
expired one-time registration token does not break container restarts.

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

## Tests

Run the persistent registration lifecycle regression test with:

```bash
bash scripts/runner/start-runner_test.sh
```
