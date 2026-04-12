import pytest

from restricted_runner.restricted_command import RestrictedCommand, RestrictedCommandError


def test_parse_resource_apply_command() -> None:
    cmd = RestrictedCommand.from_json(
        '{"operation":"resource.apply","resource":"homecloud/sites/hass","commit":"abcdef123456"}'
    )

    assert cmd.operation == "resource.apply"
    assert cmd.resource == "homecloud/sites/hass"
    assert cmd.commit == "abcdef123456"


@pytest.mark.parametrize(
    ("payload", "message"),
    [
        ('{"operation":"resource.apply","resource":"/etc/passwd","commit":"abcdef1"}', "resource path is not in allowed format"),
        ('{"operation":"resource.apply","resource":"homecloud/../secret","commit":"abcdef1"}', "resource path must not escape repository root"),
        ('{"operation":"resource.apply","resource":"homecloud/sites/hass","commit":"main"}', "commit must be a hex git sha"),
        ('{"operation":"bash","resource":"homecloud/sites/hass","commit":"abcdef1"}', "operation is invalid or not allowed"),
        ('{"operation":"resource.apply","commit":"abcdef1"}', "resource is required for operation resource.apply"),
    ],
)
def test_restricted_command_rejects_invalid_inputs(payload: str, message: str) -> None:
    with pytest.raises(RestrictedCommandError, match=message):
        RestrictedCommand.from_json(payload)


def test_repo_checkout_requires_commit_but_not_resource() -> None:
    cmd = RestrictedCommand.from_json('{"operation":"repo.checkout","commit":"abcdef1"}')

    assert cmd.operation == "repo.checkout"
    assert cmd.resource is None
    assert cmd.commit == "abcdef1"


def test_payload_must_be_json_object() -> None:
    with pytest.raises(RestrictedCommandError, match="payload must be a JSON object"):
        RestrictedCommand.from_json('[]')
