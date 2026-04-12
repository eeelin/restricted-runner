import io
import json
from pathlib import Path

from restricted_runner import restricted_dispatcher


class FakeRunner:
    def run(self, cwd: Path, args: list[str]):
        from restricted_runner.runner import CommandResult

        return CommandResult(0, "ok", "")


def test_dispatcher_dry_run_from_payload(capsys, monkeypatch, tmp_path: Path) -> None:
    monkeypatch.setattr(restricted_dispatcher, "CommandRunner", lambda: FakeRunner())

    rc = restricted_dispatcher.main(
        [
            "--repo-root",
            str(tmp_path),
            "--dry-run",
            "--payload",
            '{"operation":"resource.apply","resource":"homecloud/sites/hass","commit":"abcdef1"}',
        ]
    )

    out = capsys.readouterr().out
    payload = json.loads(out)
    assert rc == 0
    assert payload["ok"] is True
    assert payload["dry_run"] is True
    assert payload["resource"] == "homecloud/sites/hass"


def test_dispatcher_rejects_missing_payload(capsys, monkeypatch, tmp_path: Path) -> None:
    monkeypatch.setattr(restricted_dispatcher.sys, "stdin", io.StringIO(""))
    rc = restricted_dispatcher.main(["--repo-root", str(tmp_path)])

    out = capsys.readouterr().out
    payload = json.loads(out)
    assert rc == 2
    assert payload["ok"] is False
    assert payload["kind"] == "validation_error"


def test_dispatcher_payload_file(capsys, monkeypatch, tmp_path: Path) -> None:
    monkeypatch.setattr(restricted_dispatcher, "CommandRunner", lambda: FakeRunner())
    payload_file = tmp_path / "payload.json"
    payload_file.write_text(
        '{"operation":"repo.checkout","commit":"abcdef1"}',
        encoding="utf-8",
    )

    rc = restricted_dispatcher.main(["--repo-root", str(tmp_path), "--dry-run", "--payload-file", str(payload_file)])

    out = capsys.readouterr().out
    payload = json.loads(out)
    assert rc == 0
    assert payload["operation"] == "repo.checkout"
    assert payload["command"] == ["git", "checkout", "--detach", "abcdef1"]
