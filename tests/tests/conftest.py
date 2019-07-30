import os

import pytest

import lib

@pytest.fixture(autouse=True)
def run_around_tests():
    lib.clear_keys()
    lib.clear_local_config()
    # Calling yield triggers the test execution
    yield

def pytest_sessionfinish(session, exitstatus):
    # Automatically run after all tests in order to ensure that no kssh-client config files stick around
    lib.run_command("keybase fs rm /keybase/team/%s.ssh/kssh-client.config || true" % os.environ['SUBTEAM'])
    lib.run_command("keybase fs rm /keybase/team/%s.ssh.staging/kssh-client.config || true" % os.environ['SUBTEAM'])
    lib.run_command("keybase fs rm /keybase/team/%s.ssh.prod/kssh-client.config || true" % os.environ['SUBTEAM'])
    lib.run_command("keybase fs rm /keybase/team/%s.ssh.root_everywhere/kssh-client.config || true" % os.environ['SUBTEAM'])
    lib.run_command("keybase fs rm /keybase/team/%s/kssh-client.config || true" % os.environ['SUBTEAM_SECONDARY'])