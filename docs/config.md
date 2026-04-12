# restricted-runner Configuration Design

## Status

Draft

## 1. Purpose

This document defines the configuration model for `restricted-runner`.

The configuration layer is responsible for expressing:

- the execution root path
- runtime defaults
- policy allowlists
- per-script execution constraints
- caller and target restrictions
- audit-related behavior

The configuration is not just operational data. It is one of the main security boundaries of the system.

## 2. Design Goals

The configuration format should be:

- explicit
- easy to review in code review
- friendly to hand-maintained host-side files
- strict enough to support default deny behavior
- expressive enough to describe script-level restrictions without becoming a general programming language

## 3. Format Choice

The preferred configuration format is **YAML**.

Reasons:

- easier to maintain than JSON for nested host-side policy files
- readable enough for security review
- natural fit for lists of scripts and rules

JSON may still be supported as an input format later, but YAML should be the primary documented form.

## 4. Top-level Configuration Shape

A first config shape can look like this:

```yaml
version: v1

root_path: /opt/restricted-runner/root

runtime:
  default_timeout: 300s
  max_stdout_bytes: 1048576
  max_stderr_bytes: 1048576
  allow_stdin_by_default: false
  inject_runtime_env: true

audit:
  mode: stderr
  include_env_keys: false
  include_argv: true

callers:
  - id: github-actions-homecloud
    transport: ssh
    allowed_targets:
      - server
      - claw

scripts:
  - path: homecloud/site/validate
    allowed_callers:
      - github-actions-homecloud
    allowed_targets:
      - server
      - claw
    allow_argv: true
    allow_stdin: true
    allowed_env:
      - RR_TARGET
      - RR_WORKFLOW_RUN_ID
      - RR_ACTOR
    required_env:
      - RR_TARGET
    timeout: 60s

  - path: homecloud/site/apply
    allowed_callers:
      - github-actions-homecloud
    allowed_targets:
      - server
    allow_argv: true
    allow_stdin: false
    allowed_env:
      - RR_TARGET
      - RR_WORKFLOW_RUN_ID
      - RR_ACTOR
    required_env:
      - RR_TARGET
    timeout: 300s
```

## 5. Top-level Fields

### `version`

Configuration schema version.

Required.

This allows future config evolution with explicit compatibility checks.

### `root_path`

Absolute path to the execution root.

Required.

All executable scripts must be resolved under this path.

Requirements:

- must be absolute
- must not be empty
- should point to an operator-controlled directory
- should not be writable by untrusted caller identities

### `runtime`

Optional block for global execution defaults.

### `audit`

Optional block controlling audit output behavior.

### `callers`

Optional list of named caller identities.

This makes policy more reviewable than repeating raw transport identity details on every script.

### `scripts`

Required list of allowlisted executable entries.

This is the main policy surface.

## 6. Runtime Block

Example:

```yaml
runtime:
  default_timeout: 300s
  max_stdout_bytes: 1048576
  max_stderr_bytes: 1048576
  allow_stdin_by_default: false
  inject_runtime_env: true
```

### Fields

#### `default_timeout`

Default execution timeout for scripts that do not override it.

#### `max_stdout_bytes`

Maximum captured stdout size before truncation or failure behavior is applied.

#### `max_stderr_bytes`

Maximum captured stderr size before truncation or failure behavior is applied.

#### `allow_stdin_by_default`

Whether scripts may receive stdin when not explicitly configured.

Recommended default: `false`.

#### `inject_runtime_env`

Whether reserved runtime-owned environment variables should be injected automatically.

Recommended default: `true`.

## 7. Audit Block

Example:

```yaml
audit:
  mode: stderr
  include_env_keys: false
  include_argv: true
```

### Fields

#### `mode`

Possible initial values:

- `stderr`
- `jsonl-file`
- `disabled`

For the first implementation, `stderr` is likely enough.

#### `include_env_keys`

Whether audit output should include the names of caller-provided env keys.

This should default to `false` or be used carefully, to avoid accidental exposure of sensitive information.

#### `include_argv`

Whether argv should be included in audit output.

This is useful for debugging, but may need future redaction controls.

## 8. Caller Model

A caller is a named identity abstraction used by policy.

Example:

```yaml
callers:
  - id: github-actions-homecloud
    transport: ssh
    allowed_targets:
      - server
      - claw
```

### Why model callers separately

This avoids repeating transport identity details on every script entry and makes policy easier to review.

### Fields

#### `id`

Stable policy identifier for a caller.

Required.

#### `transport`

Transport type used by this caller.

Initial values might include:

- `ssh`
- `local`

#### `allowed_targets`

Optional list of logical targets this caller may address.

This does not replace per-script target restrictions. It is an outer boundary.

## 9. Script Policy Entries

Each item under `scripts` defines one allowlisted executable path under `root_path`.

Example:

```yaml
scripts:
  - path: homecloud/site/apply
    allowed_callers:
      - github-actions-homecloud
    allowed_targets:
      - server
    allow_argv: true
    allow_stdin: false
    allowed_env:
      - RR_TARGET
      - RR_WORKFLOW_RUN_ID
      - RR_ACTOR
    required_env:
      - RR_TARGET
    timeout: 300s
```

### Fields

#### `path`

Relative script path under `root_path`.

Required.

Must obey the same path validation rules as request `script` values:

- not absolute
- no `.` or `..`
- no traversal outside root

#### `allowed_callers`

List of caller ids that may invoke this script.

If omitted, the safer behavior is to treat the script as not callable.

#### `allowed_targets`

List of logical targets allowed for this script.

If request metadata includes a target, it must match this allowlist.

#### `allow_argv`

Whether arbitrary argv arrays are allowed for this script.

A later version may refine this into more structured argv validation.

#### `allow_stdin`

Whether stdin input is allowed for this script.

Recommended default: `false`.

#### `allowed_env`

Caller-provided env keys allowed for this script.

This does not include runtime-owned reserved env that the runner injects itself.

#### `required_env`

Env keys that must be present after request processing.

Useful for scripts that require target or revision context.

#### `timeout`

Optional per-script timeout override.

## 10. Reserved Runtime Environment

The runtime should own reserved env keys such as:

- `RR_REQUEST_ID`
- `RR_SCRIPT_PATH`
- `RR_ROOT_PATH`
- `RR_CALLER`
- `RR_SOURCE`
- `RR_TARGET`
- `RR_PROTOCOL_VERSION`

Configuration should not need to re-declare these as user-controlled variables.

The main policy question is whether caller-provided env keys are allowed in addition to runtime-owned values.

## 11. Matching Rules

When a request arrives, configuration matching should proceed roughly as:

1. load config and verify schema version
2. validate `root_path`
3. validate request shape
4. resolve caller identity
5. locate script policy entry by exact relative `script` path
6. check caller allowlist
7. check target allowlist
8. check argv/stdin/env constraints
9. inject runtime env
10. dispatch execution

The matching model should stay deterministic and easy to audit.

## 12. Explicit Defaults

For security review, defaults should be explicit and conservative.

Recommended behavior if a field is omitted:

- unknown top-level field: reject config
- unknown script field: reject config
- missing `root_path`: reject config
- missing `scripts`: reject config
- missing `allowed_callers`: deny all callers
- missing `allowed_targets`: deny all targets unless targetless execution is explicitly supported
- missing `allow_argv`: `false`
- missing `allow_stdin`: `false`
- missing `allowed_env`: empty list
- missing `required_env`: empty list
- missing `timeout`: inherit runtime default

## 13. Future Extensions

The initial config should stay simple, but likely future additions include:

- regex or structured argv constraints
- caller identity mapping from SSH principals or keys
- script classes or reusable policy templates
- per-script stdout/stderr size overrides
- working directory overrides
- filesystem sandbox or chroot-like controls if needed
- per-script dry-run-only mode
- result redaction policies

## 14. Security Notes

Configuration should never grow into a templating engine.

It should not support:

- arbitrary shell templates
- inline script bodies
- dynamic command interpolation
- policy expressions complex enough to hide intent

The config should remain a declarative allowlist, not a programming language.

## 15. Recommended Next Step

After this config shape is accepted, the next implementation step should be:

- define Go structs for config loading and validation
- define Go structs for request and result types
- implement path and policy matching tests
