from __future__ import annotations

import subprocess
from dataclasses import dataclass
from pathlib import Path


@dataclass(frozen=True)
class CommandResult:
    returncode: int
    stdout: str
    stderr: str


class CommandRunner:
    def run(self, cwd: Path, args: list[str]) -> CommandResult:
        proc = subprocess.run(
            args,
            cwd=str(cwd),
            check=False,
            capture_output=True,
            text=True,
        )
        return CommandResult(proc.returncode, proc.stdout, proc.stderr)


def run_command(runner: CommandRunner, cwd: Path, args: list[str]) -> CommandResult:
    return runner.run(cwd, args)
