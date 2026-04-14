# scripts

This directory contains user-facing helper scripts for local demo installation.

## Scripts

- `install-demo.sh`
- `uninstall-demo.sh`

These scripts install or remove:

- the sample SSH entrypoint wrapper
- the sample scripts under the execution root
- a rendered sample config file

They intentionally do **not** modify `authorized_keys` automatically.
That step should remain explicit and operator-reviewed.
