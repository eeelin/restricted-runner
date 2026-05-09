# docker-compose example

This example shows how to run the published GitHub Actions runner image with Docker Compose.

For the full end-to-end deployment flow, also read:

- `docs/deployment.md`

## Files

- `docker-compose.yml`
- `.env.example`

## Quick start

1. Copy the environment template:

   ```bash
   cp .env.example .env
   ```

2. Fill in:

   - `GITHUB_RUNNER_URL`
   - `GITHUB_RUNNER_TOKEN`

3. Prepare an SSH directory if your workflows use `rr-exec`:

   ```bash
   mkdir -p ssh
   chmod 700 ssh
   ```

   Put your deploy key in `ssh/`, for example:

   - `ssh/id_ed25519`
   - `ssh/id_ed25519.pub`
   - optional `ssh/config`
   - optional `ssh/known_hosts`

4. Start the runner:

   ```bash
   docker compose up -d
   ```

## Notes

- This example uses the published image `yuhuntero/restricted-runner-gha-runner:v0.0.2`.
- The SSH directory is mounted read-only into `/home/runner/.ssh`.
- The named volume `runner-work` stores the runner work directory.
- `rr-exec` is available inside the container at `/usr/local/bin/rr-exec`.
- The target host still needs the SSH forced-command setup described in `docs/deployment.md`.
