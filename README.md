# restricted-runner

Restricted host execution primitives for self-hosted runner environments.

## Current scope

This repository is being built as the standalone home for:

- restricted command parsing and validation
- host-side dispatcher entrypoints
- host-side execution wrapper logic
- structured execution results
- tests for the restricted execution path

The first consumer is HomeCloud, but the project is intentionally being split out so it can evolve more independently.

## Development

This project uses `uv`.

```bash
uv run pytest -q
```

## Current status

Early scaffold. The protocol and dispatcher are intentionally small while the repository scope is being clarified.
