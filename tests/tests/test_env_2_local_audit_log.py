import pytest

from lib import UtilitiesLib
from lib import SUBTEAM, SUBTEAM_SECONDARY, USERNAME, BOT_USERNAME, EXPECTED_HASH

class TestEnv2LocalAuditLog:
    @pytest.fixture(autouse=True, scope='class')
    def configure_env(self, test_lib):
        assert test_lib.load_env(__file__)

    @pytest.fixture(autouse=True, scope='class')
    def test_lib(self):
        return UtilitiesLib(SUBTEAM, SUBTEAM_SECONDARY, USERNAME, BOT_USERNAME, EXPECTED_HASH)

    def test_kssh(self, test_lib):
        # Test ksshing into staging as user
        with test_lib.outputs_audit_log(filename="/mnt/ca.log", expected_number=1):
            test_lib.assert_contains_hash(test_lib.run_command("""bin/kssh -q -o StrictHostKeyChecking=no user@sshd-staging "sha1sum /etc/unique" """))
