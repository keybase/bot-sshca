import pytest
from lib import TestConfig, clear_keys, clear_local_config, run_delete_kvstore_command


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
    run_delete_kvstore_command(f"{tc.subteam}.ssh")
    run_delete_kvstore_command(f"{tc.subteam}.ssh.staging")
    run_delete_kvstore_command(f"{tc.subteam}.ssh.prod")
    run_delete_kvstore_command(f"{tc.subteam}.ssh.root_everywhere")
    run_delete_kvstore_command(tc.subteam_secondary)
