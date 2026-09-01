#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
fixture_root="$(mktemp -d)"
cleanup() {
    find "${fixture_root}" -type f -delete
    find "${fixture_root}" -depth -type d -empty -delete
}
trap cleanup EXIT

make_runner_home() {
    local runner_home="$1"
    mkdir -p "${runner_home}"
    printf '%s\n' '#!/usr/bin/env bash' \
        'printf "config\\n" >> "$TEST_LOG"' \
        'printf "%s\\n" "$@" >> "$TEST_ARGS"' \
        '[[ "${TEST_CONFIG_BEHAVIOR:-success}" != fail ]] || exit 23' \
        'printf runner > .runner' \
        'printf credentials > .credentials' \
        '[[ "${TEST_CONFIG_BEHAVIOR:-success}" == missing ]] || printf rsa > .credentials_rsaparams' \
        > "${runner_home}/config.sh"
    printf '%s\n' '#!/usr/bin/env bash' \
        'printf "run\\n" >> "$TEST_LOG"' \
        'exit "${TEST_RUN_EXIT:-0}"' > "${runner_home}/run.sh"
    chmod +x "${runner_home}/config.sh" "${runner_home}/run.sh"
}

run_start_runner() {
    local runner_home="$1"
    local state_dir="$2"
    shift 2
    TEST_LOG="${fixture_root}/lifecycle.log" TEST_ARGS="${fixture_root}/config.args" \
        RUNNER_HOME="${runner_home}" RUNNER_STATE_DIR="${state_dir}" \
        GITHUB_RUNNER_URL=https://github.com/example/repository \
        GITHUB_RUNNER_TOKEN=test-token "$@" \
        bash "${repo_root}/scripts/runner/start-runner"
}

state_dir="${fixture_root}/state"
mkdir -p "${state_dir}"
for launch in first second; do
    runner_home="${fixture_root}/${launch}"
    make_runner_home "${runner_home}"
    RUNNER_NAME=custom-runner RUNNER_WORKDIR=custom-work RUNNER_LABELS=custom,label \
        run_start_runner "${runner_home}" "${state_dir}" env
done

[[ "$(grep -c '^config$' "${fixture_root}/lifecycle.log")" -eq 1 ]]
[[ "$(grep -c '^run$' "${fixture_root}/lifecycle.log")" -eq 2 ]]
grep -Fx custom-runner "${fixture_root}/config.args" >/dev/null
grep -Fx custom-work "${fixture_root}/config.args" >/dev/null
grep -Fx custom,label "${fixture_root}/config.args" >/dev/null
for state_file in .runner .credentials .credentials_rsaparams; do
    cmp "${state_dir}/${state_file}" "${fixture_root}/second/${state_file}"
    [[ "$(stat -c '%a' "${fixture_root}/second/${state_file}")" == 600 ]]
done

missing_env_home="${fixture_root}/missing-env"
make_runner_home "${missing_env_home}"
if env -u GITHUB_RUNNER_URL -u GITHUB_RUNNER_TOKEN \
    RUNNER_HOME="${missing_env_home}" RUNNER_STATE_DIR="${fixture_root}/missing-env-state" \
    bash "${repo_root}/scripts/runner/start-runner" >"${fixture_root}/missing-env.out" 2>&1; then
    printf 'start-runner unexpectedly accepted missing registration inputs\n' >&2
    exit 1
fi
grep -F 'GITHUB_RUNNER_URL is required' "${fixture_root}/missing-env.out" >/dev/null

for behavior in fail missing; do
    runner_home="${fixture_root}/${behavior}"
    scenario_state="${fixture_root}/${behavior}-state"
    make_runner_home "${runner_home}"
    if TEST_CONFIG_BEHAVIOR="${behavior}" \
        run_start_runner "${runner_home}" "${scenario_state}" env; then
        printf 'start-runner unexpectedly accepted %s registration\n' "${behavior}" >&2
        exit 1
    fi
done
[[ "$(grep -c '^run$' "${fixture_root}/lifecycle.log")" -eq 2 ]]

exit_home="${fixture_root}/exit"
make_runner_home "${exit_home}"
if TEST_RUN_EXIT=17 run_start_runner "${exit_home}" "${state_dir}" env; then
    printf 'start-runner did not preserve run.sh failure\n' >&2
    exit 1
else
    [[ "$?" -eq 17 ]]
fi

printf 'start-runner lifecycle test: PASS\n'
