import pytest

from lib import UtilitiesLib
from lib import SUBTEAM, SUBTEAM_SECONDARY, USERNAME, BOT_USERNAME, EXPECTED_HASH

@pytest.fixture(autouse=True)
def run_around_tests():
    lib = UtilitiesLib(SUBTEAM, SUBTEAM_SECONDARY, USERNAME, BOT_USERNAME, EXPECTED_HASH)
    lib.clear_keys()
    lib.clear_local_config()
    # Calling yield triggers the test execution
    yield

def pytest_sessionfinish(session, exitstatus):
    # Automatically run after all tests in order to ensure that no kssh-client config files stick around
    lib = UtilitiesLib(SUBTEAM, SUBTEAM_SECONDARY, USERNAME, BOT_USERNAME, EXPECTED_HASH)
    lib.run_command(f"keybase fs rm /keybase/team/{lib.subteam}.ssh/kssh-client.config || true" )
    lib.run_command(f"keybase fs rm /keybase/team/{lib.subteam}.ssh.staging/kssh-client.config || true" )
    lib.run_command(f"keybase fs rm /keybase/team/{lib.subteam}.ssh.prod/kssh-client.config || true" )
    lib.run_command(f"keybase fs rm /keybase/team/{lib.subteam}.ssh.root_everywhere/kssh-client.config || true" )
    lib.run_command(f"keybase fs rm /keybase/team/{lib.subteam_secondary}/kssh-client.config || true" )