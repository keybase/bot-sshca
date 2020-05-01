import pytest
from lib import (TestConfig, assert_contains_hash, load_env, outputs_audit_log,
                 run_command_with_agent)


class TestEnv2LocalAuditLog:
    @pytest.fixture(autouse=True, scope="class")
    def configure_env(self):
        assert load_env(__file__)

    @pytest.fixture(autouse=True, scope="class")
    def test_config(self):
        return TestConfig.getDefaultTestConfig()

    def test_kssh(self, test_config):
        # Test ksshing into staging as user
        with outputs_audit_log(
            test_config, filename="/shared/ca.log", expected_number=1
        ):
            assert_contains_hash(
                test_config.expected_hash,
                run_command_with_agent(
                    'bin/kssh -q -o StrictHostKeyChecking=no \
                    user@sshd-staging "sha1sum /etc/unique" '
                ),
            )
