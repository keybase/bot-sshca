import pytest
from lib import (
    TestConfig,
    assert_contains_hash,
    clear_keys,
    load_env,
    outputs_audit_log,
    run_command_with_agent,
)


class TestEnv3UserNotInFirstTeam:
    @pytest.fixture(autouse=True, scope="class")
    def configure_env(self):
        assert load_env(__file__)

    @pytest.fixture(autouse=True, scope="class")
    def test_config(self):
        return TestConfig.getDefaultTestConfig()

    def test_kssh(self, test_config):
        # Test ksshing which tests that it is correctly finding a client config
        with outputs_audit_log(
            test_config, filename="/shared/ca.log", expected_number=3
        ):
            clear_keys()
            assert_contains_hash(
                test_config.expected_hash,
                run_command_with_agent(
                    'bin/kssh -q -o StrictHostKeyChecking=no \
                    user@sshd-staging "sha1sum /etc/unique" '
                ),
            )
            clear_keys()
            assert_contains_hash(
                test_config.expected_hash,
                run_command_with_agent(
                    'bin/kssh -q -o StrictHostKeyChecking=no \
                    root@sshd-staging "sha1sum /etc/unique" '
                ),
            )
            clear_keys()
            assert_contains_hash(
                test_config.expected_hash,
                run_command_with_agent(
                    'bin/kssh -q -o StrictHostKeyChecking=no \
                    root@sshd-prod "sha1sum /etc/unique" '
                ),
            )
