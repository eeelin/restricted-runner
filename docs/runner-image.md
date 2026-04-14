# restricted-runner GitHub Actions Runner Image

## Status

Draft

## 1. Purpose

This document describes the first built-in runner image for use with GitHub Actions.

The purpose of the image is to provide a convenient execution environment that includes:

- a GitHub Actions runner base image
- SSH client tooling
- JSON tooling
- a helper shell script for invoking a remote `restricted-runner` deployment over SSH

This image does not replace `restricted-runner` itself. It is a caller-side environment.

## 2. First Version Scope

The first version includes:

- `docker/runner/Dockerfile`
- `scripts/runner/rr-exec`

The first version does not try to bundle a fully opinionated runner registration flow.
It focuses on giving workflows a consistent container image and a utility for remote dispatch.

## 3. Included Tools

The image currently installs:

- `openssh-client`
- `jq`
- `git`
- `bash`

It also installs this helper utility:

- `/usr/local/bin/rr-exec`

## 4. Current Utility Contract

### `rr-exec`

This helper:

- builds a JSON request
- sends it over SSH stdin
- expects the remote host to be configured with SSH forced-command mode
- prints the structured JSON response
- exits non-zero if `.ok != true`
- supports `--dry-run` for remote preflight when the remote SSH entrypoint is wired to honor that convention

Its `--target` flag refers to the logical policy target used by `restricted-runner`, not a filesystem path or destination directory.

This keeps GitHub workflow steps simple and consistent.

## 5. Remote Dry-run Model

The current recommended sample model is a single SSH entrypoint with an injected `RR_DRY_RUN=1` flag when preflight is desired.

This is a sample entrypoint convention, not a separate `restricted-runner` protocol field.
The SSH wrapper is responsible for mapping that environment variable to `dispatch --dry-run` on the remote host.

That keeps the SSH boundary simple:

- one entrypoint
- one helper command
- one remote command shape
- two execution modes inside the same dispatch path

## 6. Example Workflow Usage

A workflow step may look like:

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

A dry-run step may look like:

```bash
rr-exec \
  --dry-run \
  --host deploy@my-host \
  --caller github-actions-homecloud \
  --target server \
  --script homecloud/site/apply \
  --arg sites/homes/ruyi/hass \
  --env TARGET=server \
  --env ACTOR=github-actions
```

## 7. Naming Notes

### target

In this project, `target` means a logical target name used for policy matching, for example `server` or `claw`.
It does not mean a path, directory, or hostname.

## 8. Security Notes

- the runner image should use a dedicated SSH key for the remote restricted account
- that key should be limited through forced-command mode on the target host
- workflows should avoid embedding secrets directly into ad hoc shell strings
- the remote host remains the policy enforcement point

## 9. Recommended Next Steps

- tighten exactly how caller and target are injected on the SSH boundary
- add runner image build automation
- add container publishing workflow later if the image becomes part of the release contract
