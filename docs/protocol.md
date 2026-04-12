# restricted-runner Protocol Design

## Status

Draft

## 1. Protocol Direction

`restricted-runner` should be modeled as a **CGI-like execution gateway over an SSH transport**.

This means the tool should not be designed as a generic RPC framework and not as an arbitrary remote shell. Instead, it should behave like a narrow execution bridge:

- a request arrives through a trusted transport, initially SSH
- the request identifies a script or executable by **relative path** under a configured root directory
- arguments and environment are passed in a controlled, CGI-inspired structure
- stdin, stdout, and stderr remain the primary execution streams
- the runner enforces path constraints, policy checks, and audit logging before invoking the target program

This model is intentionally simple. It mirrors the strengths of CGI:

- clear process boundary
- explicit request-to-process mapping
- transport independence at the protocol layer
- easy interoperability with scripts and binaries

## 2. Core Model

On each target host, `restricted-runner` is configured with an **execution root path**.

Example:

```text
/opt/restricted-runner/root
```

Only files under that root path are eligible for execution.

A request does not provide an absolute executable path. Instead, it provides a **relative executable path**, for example:

```json
{
  "script": "homecloud/site/apply"
}
```

The final executable path is resolved as:

```text
<root_path>/<script>
```

For example:

```text
/opt/restricted-runner/root/homecloud/site/apply
```

This is one of the main security boundaries of the protocol.

## 3. Why a CGI-like Model

The CGI mental model gives us a useful shape:

- the request describes execution context
- execution happens as a spawned process
- request fields map to argv, env, and stdin
- result comes back through exit code, stdout, and stderr

That is close to what we need for a restricted SSH execution path.

Compared with a richer RPC protocol, this approach has benefits:

- easier to reason about
- easy to implement in Go
- natural fit for shell scripts and small binaries
- avoids building a heavyweight agent protocol too early
- aligns well with the desired trust boundary

## 4. Execution Root

### Definition

Each runner instance must be configured with a single execution root, for example:

```yaml
root_path: /opt/restricted-runner/root
```

### Requirements

The root path must:

- be an absolute path
- be owned and managed by the host operator
- contain only explicitly approved executables or scripts
- not be writable by untrusted callers

### Security Properties

The root path boundary is intended to ensure:

- callers cannot point execution to arbitrary host binaries
- callers cannot select files outside the approved tree
- policy decisions can be expressed in terms of relative paths under a known root

## 5. Request Shape

A first protocol request can look like this:

```json
{
  "version": "v1",
  "request_id": "6d6c3b4c-0c0e-4b62-b021-3d8dcf3a2f12",
  "script": "homecloud/site/apply",
  "argv": [
    "sites/homes/ruyi/hass",
    "--revision",
    "abcdef123456"
  ],
  "env": {
    "RR_TARGET": "server",
    "RR_WORKFLOW_RUN_ID": "1234567890",
    "RR_ACTOR": "eeelin"
  },
  "stdin": null,
  "metadata": {
    "source": "github-actions",
    "repository": "eeelin/HomeCloud"
  }
}
```

## 6. Request Fields

### `version`

Protocol version string.

This allows future format evolution without silent ambiguity.

### `request_id`

A caller-supplied correlation id used for audit and traceability.

### `script`

A relative path under the configured root path.

Examples:

- `homecloud/site/apply`
- `homecloud/site/validate`
- `ops/host/status`

The protocol must reject:

- absolute paths
- empty paths
- `.` and `..` segments
- path traversal attempts
- paths outside the configured root

### `argv`

An ordered array of arguments passed directly to the target executable.

This should map to process argv and must **not** be shell-expanded.

### `env`

A string-to-string map of environment variables passed to the process.

This is the CGI-inspired part of the model.

The environment should be explicitly bounded and filtered by policy.
The caller may propose env values, but the runner must be able to:

- reject forbidden keys
- override reserved keys
- inject system-defined audit keys

### `stdin`

Optional string or byte payload provided to the process standard input.

This is useful when a request is too large or too structured to fit comfortably in argv.

### `metadata`

Opaque metadata for correlation and auditing.

Metadata is not itself permission-granting, but it is valuable for:

- audit logs
- tracing workflow runs
- associating execution with an upstream identity

## 7. CGI-inspired Environment Model

The protocol should adopt a CGI-like environment naming convention for reserved values.

For example:

- `RR_REQUEST_ID`
- `RR_SCRIPT_PATH`
- `RR_ROOT_PATH`
- `RR_CALLER`
- `RR_SOURCE`
- `RR_TARGET`
- `RR_PROTOCOL_VERSION`

If an SSH transport is used, transport-derived fields may also be injected, for example:

- `RR_SSH_CONNECTION`
- `RR_SSH_ORIGINAL_COMMAND`

### Reserved environment rules

The caller must not be allowed to override reserved `RR_*` variables that are owned by the runtime.

The runtime may either:

- reject conflicting caller keys
- or ignore caller-provided values for reserved keys and replace them

The former is easier to reason about and preferable in the first version.

## 8. Result Shape

A result should be emitted as structured JSON, for example:

```json
{
  "ok": true,
  "request_id": "6d6c3b4c-0c0e-4b62-b021-3d8dcf3a2f12",
  "script": "homecloud/site/apply",
  "resolved_path": "/opt/restricted-runner/root/homecloud/site/apply",
  "exit_code": 0,
  "stdout": "deploy ok\n",
  "stderr": "",
  "started_at": "2026-04-12T09:30:00Z",
  "finished_at": "2026-04-12T09:30:03Z"
}
```

## 9. Policy Layer

The protocol alone is not enough. It needs a policy layer that sits between request parsing and execution.

Policy should answer questions such as:

- is this `script` allowed at all?
- is this script allowed for this caller identity or target?
- are these argv patterns allowed?
- which env keys are allowed?
- is stdin allowed for this script?
- is execution read-only, dry-run-only, or fully allowed?

A policy config may look roughly like:

```yaml
root_path: /opt/restricted-runner/root
scripts:
  - path: homecloud/site/apply
    allow_argv: true
    allow_stdin: false
    allowed_env:
      - RR_TARGET
      - RR_WORKFLOW_RUN_ID
      - RR_ACTOR
  - path: homecloud/site/validate
    allow_argv: true
    allow_stdin: true
    allowed_env:
      - RR_TARGET
```

This is only illustrative. The exact policy schema can evolve separately.

## 10. Path Resolution Rules

Path resolution must be strict.

Given:

- `root_path=/opt/restricted-runner/root`
- `script=homecloud/site/apply`

Resolution must:

1. reject absolute input paths
2. reject empty segments
3. reject `.` and `..`
4. clean the path
5. join it with the root path
6. verify that the resolved path is still under the root path
7. verify that the target exists and is executable if execution is requested

This should be treated as a first-class security invariant.

## 11. Transport Considerations

The initial transport is expected to be SSH.

That means the outer shell command may look like a forced-command or restricted-command entrypoint, for example:

```text
restricted-runner dispatch --config /etc/restricted-runner/config.yaml
```

The actual request body may be supplied through:

- stdin
- an SSH command argument containing compact JSON
- a file provided through a tightly controlled wrapper

The protocol should not depend on SSH specifics beyond optional metadata injection.
That keeps the protocol reusable if a future local socket or Unix domain socket transport is added.

## 12. Error Model

Errors should be split conceptually into at least three layers:

### Validation errors

The request itself is malformed or unsafe.

Examples:

- invalid JSON
- missing required fields
- invalid script path
- forbidden env key

### Policy errors

The request is well-formed but not allowed.

Examples:

- script not allowlisted
- argv not permitted for this script
- stdin not allowed

### Execution errors

The request passed validation and policy, but the process failed.

Examples:

- executable not found under root
- permission denied on the host
- process exit code non-zero
- timeout

## 13. Security Constraints

The protocol must enforce the following:

- no arbitrary shell interpretation
- no absolute executable paths from callers
- no path traversal outside root
- no unrestricted environment injection
- no silent mixing of policy and request data
- no execution outside approved script set

## 14. Recommended Next Steps

The next implementation documents should likely be:

- `docs/config.md` for policy configuration shape
- `docs/architecture.md` for package boundaries and runtime flow
- Go types for `Request` and `Result`
- Go path resolution and validation tests

## 15. Current Conclusion

The CGI-like model is a strong fit for `restricted-runner`.

It gives the project:

- a simple mental model
- a narrow and testable security boundary
- clear host-side deployment rules
- straightforward interoperability with scripts and binaries

Most importantly, it frames the system as a **restricted process gateway**, not as a generic remote execution layer. That is exactly the constraint we want.
