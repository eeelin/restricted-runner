from pathlib import Path

from restricted_runner.host_executor import HostExecutionRequest, execute_host_request
from restricted_runner.runner import CommandResult


class FakeRunner:
    def __init__(self, results: dict[tuple[str, str], CommandResult]) -> None:
        self.results = results
        self.calls: list[tuple[str, str]] = []

    def run(self, cwd: Path, args: list[str]) -> CommandResult:
        target = " ".join(args)
        key = (cwd.as_posix(), target)
        self.calls.append(key)
        return self.results.get(key, CommandResult(0, "", ""))


def test_checkout_repo_dry_run_returns_command(tmp_path: Path) -> None:
    result = execute_host_request(
        HostExecutionRequest(operation="repo.checkout", repo_root=tmp_path, commit="abcdef1", dry_run=True),
        runner=FakeRunner({}),
    )

    assert result["ok"] is True
    assert result["dry_run"] is True
    assert result["command"] == ["git", "checkout", "--detach", "abcdef1"]


def test_resource_validate_runs_command(tmp_path: Path) -> None:
    resource_dir = tmp_path / "homecloud/sites/hass"
    resource_dir.mkdir(parents=True)
    runner = FakeRunner({(resource_dir.as_posix(), "true"): CommandResult(0, "", "")})

    result = execute_host_request(
        HostExecutionRequest(
            operation="resource.validate",
            repo_root=tmp_path,
            resource="homecloud/sites/hass",
            commit="abcdef1",
        ),
        runner=runner,
    )

    assert result["ok"] is True
    assert result["resource"] == "homecloud/sites/hass"
    assert result["command"] == ["true"]
    assert runner.calls == [(resource_dir.as_posix(), "true")]


def test_resource_apply_returns_failure_result(tmp_path: Path) -> None:
    resource_dir = tmp_path / "homecloud/sites/hass"
    resource_dir.mkdir(parents=True)
    runner = FakeRunner({(resource_dir.as_posix(), "true"): CommandResult(2, "", "apply failed")})

    result = execute_host_request(
        HostExecutionRequest(
            operation="resource.apply",
            repo_root=tmp_path,
            resource="homecloud/sites/hass",
            commit="abcdef1",
        ),
        runner=runner,
    )

    assert result["ok"] is False
    assert result["returncode"] == 2
    assert result["stderr"] == "apply failed"


def test_resource_logs_is_non_executing_result(tmp_path: Path) -> None:
    result = execute_host_request(
        HostExecutionRequest(operation="resource.logs", repo_root=tmp_path, resource="homecloud/sites/hass", dry_run=True),
        runner=FakeRunner({}),
    )

    assert result["ok"] is True
    assert result["resource"] == "homecloud/sites/hass"
    assert "implementation-specific" in str(result["log_hint"])
