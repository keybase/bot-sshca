import pytest

from lib import assert_contains_hash, clear_keys, load_env, outputs_audit_log, run_command

class TestEnv3UserNotInFirstTeam:
    @pytest.fixture(autouse=True, scope='class')
    def configure_env(self):
        load_env(__file__)

    @outputs_audit_log(filename="/mnt/ca.log", expected_number=3)
    def test_kssh(self):
        # Test ksshing which tests that it is correctly finding a client config
        clear_keys()
        assert_contains_hash(run_command("""bin/kssh -q -o StrictHostKeyChecking=no user@sshd-staging "sha1sum /etc/unique" """))
        clear_keys()
        assert_contains_hash(run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-staging "sha1sum /etc/unique" """))
        clear_keys()
        assert_contains_hash(run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))
