import pytest

from lib import UtilitiesLib
from lib import SUBTEAM, SUBTEAM_SECONDARY, USERNAME, BOT_USERNAME, EXPECTED_HASH

class TestEnv3UserNotInFirstTeam:
    @pytest.fixture(autouse=True, scope='class')
    def configure_env(self, test_lib):
        assert test_lib.load_env(__file__)

    @pytest.fixture(autouse=True, scope='class')
    def test_lib(self):
        return UtilitiesLib(SUBTEAM, SUBTEAM_SECONDARY, USERNAME, BOT_USERNAME, EXPECTED_HASH)

    def test_kssh(self, test_lib):
        # Test ksshing which tests that it is correctly finding a client config
        with test_lib.outputs_audit_log(filename="/mnt/ca.log", expected_number=3):
            test_lib.clear_keys()
            test_lib.assert_contains_hash(test_lib.run_command("""bin/kssh -q -o StrictHostKeyChecking=no user@sshd-staging "sha1sum /etc/unique" """))
            test_lib.clear_keys()
            test_lib.assert_contains_hash(test_lib.run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-staging "sha1sum /etc/unique" """))
            test_lib.clear_keys()
            test_lib.assert_contains_hash(test_lib.run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))
