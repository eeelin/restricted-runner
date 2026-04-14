# restricted-runner Design

## Status

Draft

## 1. Project Goal

`restricted-runner` is an independent tool for constrained execution scenarios. Its goal is to provide a **verifiable, auditable, and tightly scoped host execution path** for environments that run self-hosted runners.

The core problem it solves is not simply "how to run commands remotely", but rather:

- how to avoid giving broad host privileges directly to a CI runner
- how to limit host execution to a clearly defined set of allowed operations
- how to make those operations strictly validated, auditable, and machine-readable
- how to let upstream systems, such as HomeCloud, integrate a dedicated security boundary instead of spreading that logic across application repositories

## 2. Background

In the HomeCloud deployment scenario, the plan is to move from the older timer-driven deploy agent toward a GitHub Actions self-hosted runner model.

However, giving a runner container broad host execution permissions is not acceptable. A runner may be influenced by workflow input, and if it can directly obtain high-privilege host execution power, the trust boundary becomes too weak.

Because of that, a separate restricted execution component is needed to take responsibility for:

- receiving structured execution requests
- validating whether a request is well-formed
- validating whether a request matches allowed policy
- mapping allowed requests onto a limited set of host-side actions
- returning a unified structured result
- preserving a complete audit trail

## 3. Non-goals

At this stage, this project does **not** aim to:

- implement a full GitHub Actions runner container
- own application-specific deployment policy details
- provide a general-purpose remote shell
- become an arbitrary command execution proxy
- handle the full lifecycle of a complex orchestration platform

In short, it is not a remote control framework and not a generic SSH wrapper.

## 4. Technology Stack

### Choice

The project uses **Go**, following the latest stable release line.

The current repository bootstrap target is:

- Go `1.24.0`

### Why Go

- single-binary delivery, suitable for hosts and constrained environments
- minimal runtime dependencies
- good fit for CLI tools, protocol validation, policy checks, and structured logging
- solid testing and cross-compilation support
- a better long-term fit for an independent tool than a script collection inside another repository

## 5. High-level Architecture

The system is expected to contain four boundary roles:

1. **Upstream orchestrator**
   - for example, a GitHub Actions workflow
   - expresses intent, but does not directly own host privileges

2. **Runner execution environment**
   - for example, a self-hosted runner container
   - receives workflow tasks and invokes `restricted-runner`
   - should not hold broad host privileges on its own

3. **restricted-runner**
   - this project
   - responsible for request parsing, policy validation, dispatch, result output, and audit recording

4. **Restricted host action layer**
   - the local action surface invoked by `restricted-runner`
   - must consist of explicit allowlisted actions, not arbitrary shell execution

## 6. Core Design Principles

### 6.1 Deny by default

All requests are denied by default. Only explicitly allowed operations may run.

### 6.2 Inputs must be structured

The protocol must not accept concatenated shell strings as request input.
Instead, requests should be expressed through a structured CGI-like model, with fields such as:

- `script`
- `argv`
- `env`
- `stdin`
- `metadata`

The detailed request and execution protocol is defined in `docs/protocol.md`.

### 6.3 Policy and execution are separate concerns

"Is this allowed?" and "How is this executed?" are separate layers:

- the policy layer decides whether a request is allowed
- the executor layer maps allowed requests to controlled actions

### 6.4 Results must be structured

Every execution result must be emitted as a consistent JSON structure containing at least:

- `ok`
- `operation`
- `target`
- `resource`
- `exit_code`
- `stdout`
- `stderr`
- audit-related metadata

### 6.5 Everything must be auditable

Every request should produce enough context to trace:

- who initiated it
- what was requested
- which policy rule matched
- whether it was executed or rejected
- what the final result was

## 7. Logical Package Layout

A suggested package structure is:

- `cmd/restricted-runner/`
  - CLI entrypoint
- `internal/protocol/`
  - request and response structures
  - JSON encode/decode
  - field validation
- `internal/policy/`
  - allowlists and default-deny matching
  - target, resource, and operation rules
- `internal/dispatch/`
  - request routing to handlers
- `internal/executor/`
  - controlled execution layer
  - no arbitrary shell passthrough
- `internal/audit/`
  - audit logging
- `internal/config/`
  - runtime and policy configuration loading

## 8. Protocol Model

The protocol is CGI-like and root-path-based.

At a high level:

- the target host exposes a configured execution root path
- callers select an executable by relative `script` path under that root
- arguments are passed as structured `argv`
- environment variables are passed through a filtered `env` map
- optional request input is passed through `stdin`
- execution result is returned as structured JSON

The detailed protocol model, request shape, path resolution rules, and result schema are defined in `docs/protocol.md`.

## 9. Policy Model

The first policy model should use **static configuration with default deny**.

Examples of policy concerns include:

- which relative `script` paths are allowed
- which callers or targets may invoke those scripts
- which argv patterns are allowed
- which env keys are allowed
- whether stdin is allowed for a given script

Policy configuration must not allow arbitrary shell templates.
Its job should be to:

- allow or reject requests
- select a controlled executable path under the configured root
- provide limited execution constraints

## 10. Execution Model

The first version of `restricted-runner` should execute only files located under the configured root path.

A request selects a relative script path, and the runtime resolves:

```text
<root_path>/<script>
```

Execution must:

- reject absolute paths
- reject path traversal
- reject scripts outside the configured root
- pass argv directly without shell expansion
- filter env through policy
- preserve stdin, stdout, stderr, and exit code as first-class execution streams

## 11. CLI Shape

The first CLI version can expose:

### `restricted-runner dispatch`

Consumes a JSON payload and executes the full flow:

1. parse
2. validate
3. policy check
4. dispatch
5. emit structured result

### `restricted-runner validate`

Validates the request and applicable policy without performing execution.

### `restricted-runner version`

Prints version information.

## 12. Configuration Shape

The tool should support:

- `--config <path>` to specify a configuration file
- `--payload <json>` to pass a request directly
- `--payload-file <path>` to read a request from file
- stdin input

The preferred configuration format is likely:

- YAML
- or JSON

YAML is preferred for hand-maintained policy files.

## 13. Security Requirements

### Must-have requirements

- default deny
- no arbitrary shell passthrough
- no path traversal or repository escape
- no unregistered operations
- no bypass of target or resource allowlists
- no accidental leakage of sensitive data in logs

### Later hardening goals

- persisted audit logs
- stronger target identity binding
- more detailed policy match reasons
- signed results or stronger integrity guarantees

## 14. Relationship to HomeCloud

HomeCloud is an early consumer, but `restricted-runner` should not be modeled as a HomeCloud-only tool.

That means:

- protocol fields should remain neutral where possible
- `site` should not become the only first-class concept
- HomeCloud path conventions should not be hardcoded into the protocol layer
- HomeCloud-specific resource naming and handler mapping can live in an integration layer

## 15. First-phase Deliverables

The first phase should optimize for getting the boundary right, not for full feature depth.

### Phase 1

- initialize the Go project
- define request and result protocol structures
- define policy configuration structures
- implement the minimal CLI framework
- implement parse and validate
- implement the smallest useful dispatcher
- add unit tests for protocol and policy behavior

### Phase 2

- implement static policy allowlists
- implement controlled handler registration
- implement dry-run mode
- emit structured JSON results
- add audit logging

### Phase 3

- integrate with HomeCloud for a real proof of concept
- define target, resource, and revision mapping rules
- evaluate transport and deployment shape

## 16. Current Conclusion

We should not rush into complex execution logic yet.
The first things to make solid are:

- the Go project foundation
- the design documents
- the neutral protocol model
- the boundary between policy and dispatch

Once those are clear, it becomes much safer to add real executors, SSH restricted command integration, and HomeCloud-specific adapters without repeatedly reworking the whole design.
