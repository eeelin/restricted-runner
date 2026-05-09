# restricted-runner Deployment Guide

## Status

Draft

## 1. Purpose

This guide explains how to deploy the end-to-end `restricted-runner` setup in a practical way.

It covers two sides:

- the GitHub Actions runner side
- the target host side

The intended model is:

1. a workflow runs inside a self-hosted runner container
2. the runner calls `rr-exec`
3. `rr-exec` opens SSH to a restricted account on the target host
4. SSH forced-command mode enters `restricted-runner`
5. `restricted-runner` validates policy and either preflights or executes a host-side script under the configured root

## 2. Deployment Overview

A practical deployment usually contains these pieces.

### Runner side

- the published runner image `yuhuntero/restricted-runner-gha-runner:<tag>`
- an SSH private key used only for the restricted target host account
- workflow steps that call `/usr/local/bin/rr-exec`

### Target host side

- the `restricted-runner` binary
- a config file such as `/etc/restricted-runner/config.yaml`
- an execution root such as `/opt/restricted-runner/root`
- a forced-command SSH entrypoint script
- an `sshd_config` entry that accepts the forwarded metadata env keys
- an `authorized_keys` entry with command restrictions

## 3. Runner Deployment

### Option A: Docker Compose

The quickest way to deploy the runner is to use the example under:

- `examples/docker-compose/docker-compose.yml`
- `examples/docker-compose/.env.example`

Typical steps:

1. copy `.env.example` to `.env`
2. set `GITHUB_RUNNER_URL`
3. set `GITHUB_RUNNER_TOKEN`
4. mount an SSH directory containing the deploy key
5. run `docker compose up -d`

The example image already contains:

- `ssh`
- `jq`
- `bash`
- `git`
- `/usr/local/bin/rr-exec`

### Runner SSH materials

The mounted SSH directory should usually contain:

- `id_ed25519`
- `id_ed25519.pub`
- optional `known_hosts`
- optional `config`

Keep this key dedicated to the restricted target-host account.
Do not reuse a broad administrator key.

## 4. Target Host Deployment

### 4.1 Install the binary

Install `restricted-runner` somewhere stable, for example:

```text
/usr/local/bin/restricted-runner
```

You may use the release binary from GitHub Releases.

### 4.2 Install the config file

A practical config location is:

```text
/etc/restricted-runner/config.yaml
```

The config should define:

- `root_path`
- callers
- scripts
- allowed targets
- allowed env keys
- stdin policy

### 4.3 Install the execution root

A practical root is:

```text
/opt/restricted-runner/root
```

Scripts placed there should be operator-controlled and executable.

For example:

```text
/opt/restricted-runner/root/homecloud/site/apply
/opt/restricted-runner/root/homecloud/site/validate
```

### 4.4 Install the SSH entrypoint

Install the host-side entrypoint to:

```text
/usr/local/bin/restricted-runner-ssh-entrypoint
```

This repository provides:

- `examples/ssh/restricted-runner-ssh-entrypoint`
- `scripts/install-ssh-entrypoint.sh`
- `scripts/uninstall-ssh-entrypoint.sh`

The entrypoint reads:

- `RR_CALLER`
- `RR_TARGET`
- `RR_PREFLIGHT`
- `RR_CONFIG_PATH`
- optional `RESTRICTED_RUNNER_BIN`

and maps them into:

- `restricted-runner dispatch ...`
- or `restricted-runner dispatch --dry-run ...`

## 5. SSH Configuration

### 5.1 Accept forwarded environment variables

Your SSH server must allow the metadata env keys used by `rr-exec`.

Add this to `sshd_config`:

```text
AcceptEnv RR_CALLER RR_TARGET RR_PREFLIGHT
```

Then reload or restart sshd.

### 5.2 Restricted account

Create or choose a dedicated restricted account for forced-command access.

The account should:

- not be used for normal shell administration
- hold only the minimum SSH access needed for this flow
- use a tightly controlled `authorized_keys`

### 5.3 authorized_keys

Use a forced-command entry similar to:

```text
command="env RR_CONFIG_PATH=/etc/restricted-runner/config.yaml /usr/local/bin/restricted-runner-ssh-entrypoint",no-port-forwarding,no-agent-forwarding,no-X11-forwarding,no-pty ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA... runner-homecloud
```

Recommended restrictions:

- `command=...`
- `no-port-forwarding`
- `no-agent-forwarding`
- `no-X11-forwarding`
- `no-pty`

## 6. Workflow Usage

A typical workflow step may look like this:

```bash
rr-exec \
  --host deploy@my-host \
  --identity /home/runner/.ssh/id_ed25519 \
  --ssh-option StrictHostKeyChecking=yes \
  --caller github-actions-homecloud \
  --target server \
  --script homecloud/site/apply \
  --arg sites/homes/ruyi/hass \
  --env TARGET=server \
  --env ACTOR=github-actions
```

A preflight step may look like this:

```bash
rr-exec \
  --preflight \
  --host deploy@my-host \
  --identity /home/runner/.ssh/id_ed25519 \
  --ssh-option StrictHostKeyChecking=yes \
  --caller github-actions-homecloud \
  --target server \
  --script homecloud/site/apply \
  --arg sites/homes/ruyi/hass \
  --env TARGET=server \
  --env ACTOR=github-actions
```

## 7. Operational Notes

### target

`target` is a logical policy target name, not a path and not a hostname.
Examples may include:

- `server`
- `claw`

### preflight

`--preflight` means:

- request validation
- config load
- policy match
- path resolution
- executable preflight
- no actual execution

### Error handling

`rr-exec` prints structured remote error output when the remote task fails and exits non-zero.
This is intended to make GitHub Actions job logs useful during failure.

## 8. Security Checklist

Before using this in production, verify at least the following:

- the runner uses a dedicated SSH key
- the target host account is dedicated and restricted
- `authorized_keys` uses forced-command restrictions
- `sshd_config` only accepts the intended env keys
- the execution root is operator-managed
- scripts under `root_path` are reviewed and executable
- policy config is deny-by-default and narrowly scoped

## 9. Related Files

- `docs/runner-image.md`
- `docs/ssh.md`
- `examples/docker-compose/README.md`
- `examples/ssh/restricted-runner-ssh-entrypoint`
- `scripts/install-ssh-entrypoint.sh`
- `scripts/runner/rr-exec`
