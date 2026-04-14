# restricted-runner GitHub Actions Runner Image

## Status

Draft

## 1. Purpose

This document describes the first built-in runner image for use with GitHub Actions.

The purpose of the image is to provide a convenient execution environment that includes:

- a GitHub Actions runner base image
- SSH client tooling
- JSON tooling
- helper shell scripts for invoking a remote `restricted-runner` deployment over SSH

This image does not replace `restricted-runner` itself. It is a caller-side environment.

## 2. First Version Scope

The first version includes:

- `docker/runner/Dockerfile`
- `scripts/runner/rr-dispatch-ssh`

The first version does not try to bundle a fully opinionated runner registration flow.
It focuses on giving workflows a consistent container image and a utility for remote dispatch.

## 3. Included Tools

The image currently installs:

- `openssh-client`
- `jq`
- `git`
- `bash`

It also installs these helper utilities:

- `/usr/local/bin/rr-dispatch-ssh`
- `/usr/local/bin/rr-validate-ssh`

## 4. Current Utility Contract

### `rr-dispatch-ssh`

This helper:

- builds a JSON request
- sends it over SSH stdin
- expects the remote host to be configured with SSH forced-command mode
- prints the structured JSON response
- exits non-zero if `.ok != true`

This keeps GitHub workflow steps simple and consistent.

## 5. Validate-mode note

The current SSH forced-command design is dispatch-oriented.
A true remote validate path should be introduced carefully so it does not conflict with the forced-command boundary.

Because of that, `rr-validate-ssh` should currently be treated as provisional helper surface, not a stable contract.
A future revision should likely choose one of these approaches:

- a dedicated validate-only forced-command entrypoint
- a separate key mapped to validate mode
- a tightly constrained accepted command grammar for the SSH wrapper

## 6. Example Workflow Usage

A workflow step may look like:

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

## 7. Security Notes

- the runner image should use a dedicated SSH key for the remote restricted account
- that key should be limited through forced-command mode on the target host
- workflows should avoid embedding secrets directly into ad hoc shell strings
- the remote host remains the policy enforcement point

## 8. Recommended Next Steps

- decide the exact validate-over-SSH shape
- add runner image build automation
- add container publishing workflow later if the image becomes part of the release contract
