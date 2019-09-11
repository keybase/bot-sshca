import json
import subprocess

import pytest
import requests

from lib import TestConfig, load_env, assert_contains_hash, run_command_with_agent

class TestEnv5TwoMan:
    @pytest.fixture(autouse=True, scope='class')
    def configure_env(self):
        assert load_env(__file__)

    @pytest.fixture(autouse=True, scope='class')
    def test_config(self):
        return TestConfig.getDefaultTestConfig()

    def test_kssh_with_two_man(self, test_config):
        assert requests.get('http://autoapprover:8080/start').content == b"OK"

        try:
            assert_contains_hash(test_config.expected_hash, run_command_with_agent(f"bin/kssh --request-realm {test_config.subteam}.ssh.root_everywhere -q -o StrictHostKeyChecking=no root@sshd-prod 'sha1sum /etc/unique'"))
        finally:
            assert requests.get('http://autoapprover:8080/stop').content == b"OK"
