# restricted-runner GitHub Actions Runner Image

## Status

Draft

## 1. Purpose

This document describes the first built-in runner image for use with GitHub Actions.
For the full end-to-end deployment procedure, see `docs/deployment.md`.

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
- `scripts/install-ssh-entrypoint.sh`
- `scripts/uninstall-ssh-entrypoint.sh`
- `examples/ssh/restricted-runner-ssh-entrypoint`

The first version does not try to bundle a fully opinionated runner registration flow.
It focuses on giving workflows a consistent container image plus the SSH entrypoint wiring needed for remote execution and preflight.

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
- supports `--preflight` for remote dispatch preflight without real execution

Its `--target` flag refers to the logical policy target used by `restricted-runner`, not a filesystem path or destination directory.

This keeps GitHub workflow steps simple and consistent.

## 5. Remote Preflight Model

The recommended model is a single SSH entrypoint with `RR_CALLER`, `RR_TARGET`, and optional `RR_PREFLIGHT=1` passed through SSH environment forwarding.

This is part of the runner-side SSH integration contract for this repository.
The SSH entrypoint maps `RR_PREFLIGHT=1` to `restricted-runner dispatch --dry-run` on the remote host.
The SSH server should allow this with:

```text
AcceptEnv RR_CALLER RR_TARGET RR_PREFLIGHT
```

That keeps the SSH boundary simple:

- one entrypoint
- one helper command
- no SSH_ORIGINAL_COMMAND dependency
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

A preflight step may look like:

```bash
rr-exec \
  --preflight \
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

## 8. SSH Entrypoint Support

The host-side SSH configuration should preserve the forwarded environment variables needed by the entrypoint.
That means `sshd_config` should allow:

```text
AcceptEnv RR_CALLER RR_TARGET RR_PREFLIGHT
```


The repository now includes a host-side SSH entrypoint and install helpers:

- `examples/ssh/restricted-runner-ssh-entrypoint`
- `scripts/install-ssh-entrypoint.sh`
- `scripts/uninstall-ssh-entrypoint.sh`

These are intended to support the runner-side `rr-exec` contract directly.
The entrypoint defaults to `/usr/local/bin/restricted-runner` and may be overridden with `RESTRICTED_RUNNER_BIN` when needed for local testing or non-standard installation layouts.

## 9. Security Notes

- the runner image should use a dedicated SSH key for the remote restricted account
- that key should be limited through forced-command mode on the target host
- workflows should avoid embedding secrets directly into ad hoc shell strings
- the remote host remains the policy enforcement point

## 10. Recommended Next Steps

- consider moving caller and target trust fully to host-side key mapping when deployment constraints require stricter provenance
