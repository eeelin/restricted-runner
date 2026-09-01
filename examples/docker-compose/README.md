# docker-compose example

This example builds and runs the GitHub Actions runner image with Docker Compose.

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

   Later container restarts reuse the saved registration:

   ```bash
   docker compose restart
   ```

## Notes

- This example builds `docker/runner/Dockerfile` as
  `restricted-runner-gha-runner:local`, ensuring the Compose command and image
  stay in sync.
- The SSH directory is mounted read-only into `/home/runner/.ssh`.
- The named volume `runner-work` stores the runner work directory.
- The named volume `runner-state` stores the registered runner identity and
  credentials. Keep this volume across container restarts and image upgrades.
- `GITHUB_RUNNER_TOKEN` is only used when `runner-state` is empty. GitHub
  registration tokens expire after one hour, so generate a new token if the
  state volume is removed and the runner must be registered again.
- Avoid `docker compose down -v` during routine updates because it removes the
  `runner-state` volume. If that volume is removed, delete the stale runner in
  GitHub when needed, generate a new registration token, and start again.
- `rr-exec` is available inside the container at `/usr/local/bin/rr-exec`.
- The target host still needs the SSH forced-command setup described in `docs/deployment.md`.
