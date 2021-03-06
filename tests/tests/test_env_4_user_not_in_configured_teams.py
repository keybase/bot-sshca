import json
import subprocess

import pytest
from lib import (
    TestConfig,
    load_env,
    outputs_audit_log,
    run_command_with_agent,
    run_put_kvstore_command,
)


class TestEnv4UserNotInConfiguredTeams:
    @pytest.fixture(autouse=True, scope="class")
    def configure_env(self):
        assert load_env(__file__)

    @pytest.fixture(autouse=True, scope="class")
    def test_config(self):
        return TestConfig.getDefaultTestConfig()

    def test_kssh_no_configs(self, test_config):
        # Test that it can't find any configs
        with outputs_audit_log(
            test_config, filename="/shared/ca.log", expected_number=0
        ):
            for s in [
                "user@sshd-staging",
                "root@sshd-staging",
                "user@sshd-prod",
                "root@sshd-prod",
            ]:
                try:
                    run_command_with_agent(
                        f"bin/kssh -q -o StrictHostKeyChecking=no {s} "
                        f'"sha1sum /etc/unique" '
                    )
                    assert False
                except subprocess.CalledProcessError as e:
                    assert b"Did not find any configs" in e.output

    def test_kssh_spoofed_config(self, test_config):
        # Test that even when kssh is forced to run by a spoofed config, the CA
        # bot ignores messages that are in the wrong channel
        with outputs_audit_log(
            test_config, filename="/shared/ca.log", expected_number=0
        ):
            client_config = json.dumps(
                {
                    "teamname": f"{test_config.subteam}.ssh",
                    "channelname": "",
                    "botname": test_config.bot_username,
                }
            )
            run_put_kvstore_command(f"{test_config.subteam}.ssh", client_config)

            for s in [
                "user@sshd-staging",
                "root@sshd-staging",
                "user@sshd-prod",
                "root@sshd-prod",
            ]:
                try:
                    run_command_with_agent(
                        f"bin/kssh -q -o StrictHostKeyChecking=no {s} "
                        f'"sha1sum /etc/unique" '
                    )
                    assert False
                except subprocess.CalledProcessError as e:
                    assert (
                        b"Failed to get a signed key from the CA: "
                        b"timed out while waiting for a response from the CA"
                    ) in e.output
