import pytest
from lib import TestConfig, clear_keys, clear_local_config, run_command


@pytest.fixture(autouse=True)
def run_around_tests():
    clear_keys()
    clear_local_config()
    # Calling yield triggers the test execution
    yield


def pytest_sessionfinish(session, exitstatus):
    # Automatically run after all tests in order to ensure that no kssh-client
    # configs stick around
    tc = TestConfig.getDefaultTestConfig()
    run_command(
        f'echo \'{{"method": "del", '
        f'"params": {{"options": {{"team": '
        f'"{tc.subteam}.ssh", "namespace": "__sshca", '
        f'"entryKey": "kssh_config"}}}}}}\' | '
        f"xargs -0 -I del keybase kvstore api -m del || true"
    )
    run_command(
        f'echo \'{{"method": "del", '
        f'"params": {{"options": {{"team": '
        f'"{tc.subteam}.ssh.staging", "namespace": "__sshca", '
        f'"entryKey": "kssh_config"}}}}}}\' | '
        f"xargs -0 -I del keybase kvstore api -m del || true"
    )
    run_command(
        f'echo \'{{"method": "del", '
        f'"params": {{"options": {{"team": '
        f'"{tc.subteam}.ssh.prod", "namespace": "__sshca", '
        f'"entryKey": "kssh_config"}}}}}}\' | '
        f"xargs -0 -I del keybase kvstore api -m del || true"
    )
    run_command(
        f'echo \'{{"method": "del", '
        f'"params": {{"options": {{"team": '
        f'"{tc.subteam}.ssh.root_everywhere", "namespace": "__sshca", '
        f'"entryKey": "kssh_config"}}}}}}\' | '
        f"xargs -0 -I del keybase kvstore api -m del || true"
    )
    run_command(
        f'echo \'{{"method": "del", '
        f'"params": {{"options": {{"team": '
        f'"{tc.subteam_secondary}", "namespace": "__sshca", '
        f'"entryKey": "kssh_config"}}}}}}\' | '
        f"xargs -0 -I del keybase kvstore api -m del || true"
    )
