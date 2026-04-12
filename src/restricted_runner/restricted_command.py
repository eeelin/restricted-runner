from __future__ import annotations

from dataclasses import dataclass
from pathlib import PurePosixPath
import json
import re

RESOURCE_RE = re.compile(r"^[a-zA-Z0-9._/-]+$")
COMMIT_SHA_RE = re.compile(r"^[0-9a-f]{7,40}$")
OPERATION_VALUES = {
    "repo.checkout",
    "resource.validate",
    "resource.apply",
    "resource.logs",
    "resource.status",
}


class RestrictedCommandError(ValueError):
    pass


@dataclass(frozen=True)
class RestrictedCommand:
    operation: str
    resource: str | None = None
    commit: str | None = None

    @classmethod
    def from_json(cls, payload: str) -> "RestrictedCommand":
        try:
            data = json.loads(payload)
        except json.JSONDecodeError as exc:
            raise RestrictedCommandError("invalid JSON payload") from exc
        if not isinstance(data, dict):
            raise RestrictedCommandError("payload must be a JSON object")
        return cls(
            operation=_require_operation(data.get("operation")),
            resource=_normalize_resource(data.get("resource")),
            commit=_normalize_commit(data.get("commit")),
        ).validate()

    def validate(self) -> "RestrictedCommand":
        if self.operation in {"resource.validate", "resource.apply", "resource.logs", "resource.status"}:
            if not self.resource:
                raise RestrictedCommandError(f"resource is required for operation {self.operation}")
        if self.operation in {"repo.checkout", "resource.validate", "resource.apply"}:
            if not self.commit:
                raise RestrictedCommandError(f"commit is required for operation {self.operation}")
        return self


def _require_operation(value: object) -> str:
    if not isinstance(value, str) or value not in OPERATION_VALUES:
        raise RestrictedCommandError("operation is invalid or not allowed")
    return value


def _normalize_resource(value: object) -> str | None:
    if value is None:
        return None
    if not isinstance(value, str):
        raise RestrictedCommandError("resource must be a string")
    raw = value.strip()
    if not raw or raw.startswith("/"):
        raise RestrictedCommandError("resource path is not in allowed format")
    normalized = raw.strip("/")
    path = PurePosixPath(normalized)
    if path.is_absolute() or ".." in path.parts or "." in path.parts:
        raise RestrictedCommandError("resource path must not escape repository root")
    if not normalized or not RESOURCE_RE.fullmatch(normalized):
        raise RestrictedCommandError("resource path is not in allowed format")
    return normalized


def _normalize_commit(value: object) -> str | None:
    if value is None:
        return None
    if not isinstance(value, str):
        raise RestrictedCommandError("commit must be a string")
    normalized = value.strip().lower()
    if not COMMIT_SHA_RE.fullmatch(normalized):
        raise RestrictedCommandError("commit must be a hex git sha")
    return normalized
