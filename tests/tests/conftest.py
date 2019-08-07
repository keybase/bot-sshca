import pytest

from lib import TestConfig, run_command, clear_keys, clear_local_config

@pytest.fixture(autouse=True)
def run_around_tests():
    clear_keys()
    clear_local_config()
    # Calling yield triggers the test execution
    yield

def pytest_sessionfinish(session, exitstatus):
    # Automatically run after all tests in order to ensure that no kssh-client config files stick around
    tc = TestConfig.getDefaultTestConfig()
    run_command(f"keybase fs rm /keybase/team/{tc.subteam}.ssh/kssh-client.config || true" )
    run_command(f"keybase fs rm /keybase/team/{tc.subteam}.ssh.staging/kssh-client.config || true" )
    run_command(f"keybase fs rm /keybase/team/{tc.subteam}.ssh.prod/kssh-client.config || true" )
    run_command(f"keybase fs rm /keybase/team/{tc.subteam}.ssh.root_everywhere/kssh-client.config || true" )
    run_command(f"keybase fs rm /keybase/team/{tc.subteam_secondary}/kssh-client.config || true" )