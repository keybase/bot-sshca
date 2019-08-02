import pytest

from lib import assert_contains_hash, load_env, outputs_audit_log, run_command

class TestEnv2LocalAuditLog:
    @pytest.fixture(autouse=True, scope='class')
    def configure_env(self):
        assert load_env(__file__)

    @outputs_audit_log(filename="/mnt/ca.log", expected_number=1)
    def test_kssh(self):
        # Test ksshing into staging as user
        assert_contains_hash(run_command("""bin/kssh -q -o StrictHostKeyChecking=no user@sshd-staging "sha1sum /etc/unique" """))
