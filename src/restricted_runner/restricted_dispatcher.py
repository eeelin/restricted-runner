from __future__ import annotations

import argparse
import sys
from pathlib import Path

from .host_executor import HostExecutionRequest, execute_host_request, format_host_result
from .restricted_command import RestrictedCommand, RestrictedCommandError
from .runner import CommandRunner


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="restricted-runner-dispatch")
    parser.add_argument("--repo-root", default=".")
    parser.add_argument("--payload", help="JSON payload for restricted command")
    parser.add_argument("--payload-file", help="Read JSON payload from a file")
    parser.add_argument("--dry-run", action="store_true")
    return parser


def _read_payload(args: argparse.Namespace) -> str:
    if args.payload and args.payload_file:
        raise RestrictedCommandError("use either --payload or --payload-file, not both")
    if args.payload:
        return args.payload
    if args.payload_file:
        return Path(args.payload_file).read_text(encoding="utf-8")
    data = sys.stdin.read()
    if not data.strip():
        raise RestrictedCommandError("missing restricted command payload")
    return data


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)

    try:
        payload = _read_payload(args)
        command = RestrictedCommand.from_json(payload)
        request = HostExecutionRequest(
            operation=command.operation,
            repo_root=Path(args.repo_root).resolve(),
            resource=command.resource,
            commit=command.commit,
            dry_run=args.dry_run,
        )
        result = execute_host_request(request, runner=CommandRunner())
        print(format_host_result(result))
        return 0 if result.get("ok") else 1
    except RestrictedCommandError as exc:
        print(format_host_result({"ok": False, "error": str(exc), "kind": "validation_error"}))
        return 2
    except Exception as exc:  # noqa: BLE001
        print(format_host_result({"ok": False, "error": str(exc), "kind": "execution_error"}))
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
