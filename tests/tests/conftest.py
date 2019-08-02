import os

import pytest

import lib
from lib import SUBTEAM, SUBTEAM_SECONDARY

@pytest.fixture(autouse=True)
def run_around_tests():
    lib.clear_keys()
    lib.clear_local_config()
    # Calling yield triggers the test execution
    yield

def pytest_sessionfinish(session, exitstatus):
    # Automatically run after all tests in order to ensure that no kssh-client config files stick around
    lib.run_command(f"keybase fs rm /keybase/team/{SUBTEAM}.ssh/kssh-client.config || true" )
    lib.run_command(f"keybase fs rm /keybase/team/{SUBTEAM}.ssh.staging/kssh-client.config || true" )
    lib.run_command(f"keybase fs rm /keybase/team/{SUBTEAM}.ssh.prod/kssh-client.config || true" )
    lib.run_command(f"keybase fs rm /keybase/team/{SUBTEAM}.ssh.root_everywhere/kssh-client.config || true" )
    lib.run_command(f"keybase fs rm /keybase/team/{SUBTEAM_SECONDARY}/kssh-client.config || true" )