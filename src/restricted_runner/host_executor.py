from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
import json

from .runner import CommandResult, run_command


@dataclass(frozen=True)
class HostExecutionRequest:
    operation: str
    repo_root: Path
    resource: str | None = None
    commit: str | None = None
    dry_run: bool = False


def execute_host_request(request: HostExecutionRequest, runner) -> dict[str, object]:
    if request.operation == "repo.checkout":
        return _checkout_repo(request, runner)
    if request.operation == "resource.validate":
        return _resource_command(request, runner, ["true"])
    if request.operation == "resource.apply":
        return _resource_command(request, runner, ["true"])
    if request.operation == "resource.logs":
        return {
            "ok": True,
            "operation": request.operation,
            "resource": request.resource,
            "log_hint": f"logs for {request.resource} are implementation-specific",
            "dry_run": request.dry_run,
        }
    if request.operation == "resource.status":
        return {
            "ok": True,
            "operation": request.operation,
            "resource": request.resource,
            "status": "unknown",
            "dry_run": request.dry_run,
        }
    raise ValueError(f"unsupported operation: {request.operation}")


def _checkout_repo(request: HostExecutionRequest, runner) -> dict[str, object]:
    command = ["git", "checkout", "--detach", request.commit]
    if request.dry_run:
        return {
            "ok": True,
            "operation": request.operation,
            "repo_root": str(request.repo_root),
            "commit": request.commit,
            "dry_run": True,
            "command": command,
        }
    result = run_command(runner, request.repo_root, command)
    return _command_result_payload(request, result, command)


def _resource_command(request: HostExecutionRequest, runner, command: list[str]) -> dict[str, object]:
    cwd = request.repo_root / (request.resource or "")
    if request.dry_run:
        return {
            "ok": True,
            "operation": request.operation,
            "resource": request.resource,
            "commit": request.commit,
            "dry_run": True,
            "cwd": str(cwd),
            "command": command,
        }
    result = run_command(runner, cwd, command)
    return _command_result_payload(request, result, command, cwd=cwd, extra={"resource": request.resource})


def _command_result_payload(
    request: HostExecutionRequest,
    result: CommandResult,
    command: list[str],
    cwd: Path | None = None,
    extra: dict[str, object] | None = None,
) -> dict[str, object]:
    payload: dict[str, object] = {
        "ok": result.returncode == 0,
        "operation": request.operation,
        "commit": request.commit,
        "command": command,
        "returncode": result.returncode,
        "stdout": result.stdout,
        "stderr": result.stderr,
        "dry_run": request.dry_run,
    }
    if cwd is not None:
        payload["cwd"] = str(cwd)
    if extra:
        payload.update(extra)
    return payload


def format_host_result(payload: dict[str, object]) -> str:
    return json.dumps(payload, indent=2, sort_keys=True)
