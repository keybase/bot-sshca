import pytest
import requests

from lib import assert_contains_hash, outputs_audit_log, run_command

class TestEnv2:
    @pytest.fixture(autouse=True, scope='class')
    def configure_env(self):
        assert requests.get("http://ca-bot:8080/load_env?filename=env-2-log-to-fs").content == b"OK"

    @outputs_audit_log(filename="/mnt/ca.log", expected_number=1)
    def test_kssh(self):
        # Test ksshing into staging as user
        assert_contains_hash(run_command("""bin/kssh -q -o StrictHostKeyChecking=no user@sshd-staging "sha1sum /etc/unique" """))
