# Restricted Runner Project Context

## Purpose

`restricted-runner` is being split out as a more independent tool instead of continuing to live only as an implementation detail inside the HomeCloud repository.

The intended direction is a reusable building block for running a GitHub Actions self-hosted runner in a constrained environment, with host-side execution mediated through a restricted command interface.

## Why this project exists

In HomeCloud, we want to move site deployment from the older timer-driven private-network deploy agent toward a GitHub Actions self-hosted runner model.

However, we do **not** want to grant the runner broad host privileges.
The security goal is to keep the runner contained and force any host-impacting operation through a tightly validated restricted execution path.

That naturally suggests a standalone tool with clearer responsibilities and sharper boundaries.

## Current architectural direction

The current desired shape is:

- GitHub Actions remains the orchestration and approval control plane
- a self-hosted runner runs inside a hardened container on approved hosts
- outbound network access from that container is allowlisted
- the runner container does not get unrestricted host execution permissions
- host-side actions are exposed only through SSH Restricted Commands or an equivalent restricted execution protocol
- the restricted interface accepts only a bounded command vocabulary
- command inputs are validated strictly, including site path, target host, and commit identity where applicable
- all execution should be auditable and produce structured machine-readable results

## Existing work already done in HomeCloud

Some early scaffolding has already been created in the HomeCloud repository under `tools/python/site/deploy/` and `specs/`.
This new repository should treat that work as upstream context, not as the final project shape.

### Artifacts currently created in HomeCloud

- `specs/site-deploy-runner-container-design.md`
- `specs/site-deploy-runner-container-tasks.md`
- `specs/adr-0001-site-deploy-runner-architecture.md`
- `tools/python/site/deploy/restricted_command.py`
- `tools/python/site/deploy/restricted_dispatcher.py`
- `tools/python/site/deploy/host_executor.py`
- tests for the above scaffolding

### What those HomeCloud artifacts currently cover

- architecture direction for a runner-container-based deploy path
- phased migration away from the old poll-based deploy agent
- a minimal restricted command protocol with input validation
- a dispatcher CLI scaffold
- a host executor scaffold with dry-run and structured results

## Likely scope for this repository

This repository will probably own some or all of the following:

- restricted command schema and protocol
- host-side dispatcher implementation
- host-side executor or wrapper implementation
- policy and validation logic for allowed operations
- structured result formats
- deployment or packaging assets for running the restricted component
- possibly runner container integration assets if that boundary belongs here

## Things still open

These are not decided yet:

- whether this repo should include the runner container image itself, or only the restricted execution side
- whether the restricted execution transport should remain SSH-focused or support other local transports too
- how much HomeCloud-specific site logic should stay in HomeCloud versus move into this repo as reusable primitives
- whether this tool is generic across multiple projects or still semi-opinionated toward HomeCloud

## Suggested near-term milestones

1. define repo purpose and boundaries clearly
2. port or redesign the minimal restricted-command protocol from HomeCloud
3. implement a small validated dispatcher executable
4. implement a host executor with a narrow operation allowlist
5. add tests for parser, validator, and execution result structure
6. decide packaging model and host deployment model
7. integrate back into HomeCloud as a consumer instead of embedding all logic there

## Handoff note

This repository was prepared from an OpenClaw session with eeelin on 2026-04-12 in Discord `#infra`.
The plan is to continue implementation work in a new dedicated Discord thread focused on `restricted-runner`.

When resuming, start by confirming:

- desired repo scope
- preferred implementation language and packaging
- whether HomeCloud remains the first consumer or only one of several consumers
