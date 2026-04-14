# restricted-runner SSH Forced-Command Integration

## Status

Draft

## 1. Purpose

This document describes how `restricted-runner` should be integrated with SSH forced-command mode.

The goal is to use SSH as a narrow transport and identity boundary, while keeping request validation, policy checks, path resolution, and execution logic inside `restricted-runner` itself.

## 2. Design Principle

SSH forced-command should be treated as a **gate**, not as the place where business logic lives.

That means:

- SSH is responsible for key-based identity and command restriction
- a forced command is responsible for entering a constrained program path
- `restricted-runner` is responsible for request parsing, policy checks, and execution

This keeps the system easier to audit and avoids hiding policy logic inside shell glue.

## 3. Recommended Shape

The recommended first shape is:

1. A dedicated SSH key is created for the caller, for example a runner host or runner container.
2. That key is added to `authorized_keys` with a forced command.
3. The forced command launches a small entrypoint wrapper.
4. The wrapper injects trusted caller metadata and execs `restricted-runner dispatch --config ...`.
5. The request body is passed through stdin.

## 4. Why stdin is preferred

The request body should preferably be sent through stdin, not embedded in the SSH command line.

Reasons:

- avoids shell escaping complexity
- avoids accidental logging of full payloads in process lists or shell history
- aligns with the CGI-like request model
- makes larger request bodies easier to pass safely

## 5. authorized_keys Example

A typical `authorized_keys` entry may look like:

```text
command="/usr/local/bin/restricted-runner-ssh-entrypoint",no-port-forwarding,no-agent-forwarding,no-X11-forwarding,no-pty ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA... runner-homecloud
```

Recommended restrictions:

- `command=...`
- `no-port-forwarding`
- `no-agent-forwarding`
- `no-X11-forwarding`
- `no-pty`

Optional later restrictions may include source IP constraints where stable networking allows it.

## 6. Entrypoint Wrapper Role

The wrapper should do only a small amount of SSH-specific handling:

- capture transport-derived metadata such as `SSH_CONNECTION`
- reject unexpected `SSH_ORIGINAL_COMMAND` usage if the deployment forbids it
- inject a trusted caller id into the environment
- exec `restricted-runner dispatch --config ...`

It should **not**:

- implement business policy
- rewrite requests in complex ways
- build arbitrary shell commands
- bypass `restricted-runner` validation or policy layers

## 7. Example Entrypoint Behavior

A minimal wrapper can behave like this:

```bash
export RR_CALLER=github-actions-homecloud
export RR_TRANSPORT=ssh
exec /usr/local/bin/restricted-runner dispatch --config /etc/restricted-runner/config.yaml
```

A stricter wrapper may also reject any non-empty `SSH_ORIGINAL_COMMAND` depending on the transport shape in use.

## 8. Caller Identity Strategy

The caller identity should not be trusted if it comes from the request payload alone.

Preferred order of trust:

1. SSH key identity or key-to-caller mapping
2. forced-command wrapper injected environment
3. request payload metadata for audit only

This means the wrapper or SSH integration layer should provide the trusted caller identity to `restricted-runner`.

## 9. Target Identity Strategy

For many deployments, the target host identity should also be considered trusted host-local configuration, not caller-controlled input.

For example, a host-local wrapper may inject:

- `RR_TARGET=server`

That reduces the risk of a caller connecting to one host while claiming to target another.

## 10. Runtime Flow

A typical runtime flow is:

1. caller opens SSH connection using a restricted key
2. SSH applies forced-command restrictions
3. forced-command wrapper starts
4. wrapper injects trusted caller and transport metadata
5. wrapper execs `restricted-runner dispatch --config /etc/restricted-runner/config.yaml`
6. request JSON is read from stdin
7. `restricted-runner` performs protocol validation
8. `restricted-runner` performs policy matching
9. `restricted-runner` resolves the script path under `root_path`
10. `restricted-runner` performs preflight or execution
11. structured JSON result is written to stdout

## 11. Installation Layout Suggestion

A practical first layout on the target host could be:

```text
/usr/local/bin/restricted-runner
/usr/local/bin/restricted-runner-ssh-entrypoint
/etc/restricted-runner/config.yaml
/opt/restricted-runner/root/
/home/<restricted-user>/.ssh/authorized_keys
```

## 12. Security Notes

### Keep wrappers small

The entrypoint wrapper should stay intentionally small. If it becomes large, policy logic is likely leaking out of the application boundary.

### Avoid passing payload in command line arguments

stdin is preferred for request transport.

### Avoid broad user privileges

The SSH account used for forced-command mode should be dedicated and restricted as much as possible.

### Keep execution root operator-managed

The scripts under `root_path` should be operator-controlled and reviewed.

## 13. Recommended Next Steps

- add a sample wrapper under `examples/ssh/`
- add an `authorized_keys` example
- add install and uninstall scripts for a local demonstration deployment
- later, teach the Go binary to consume trusted caller and target metadata from environment variables instead of requiring them only as CLI flags
