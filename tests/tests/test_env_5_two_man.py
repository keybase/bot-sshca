import json
import subprocess

import pytest
import requests

from lib import TestConfig, load_env, assert_contains_hash, run_command_with_agent, outputs_audit_log

from contextlib import contextmanager

@contextmanager
def autoapprover():
    assert requests.get('http://autoapprover:8080/start').content == b"OK"
    yield
    assert requests.get('http://autoapprover:8080/stop').content == b"OK"

class TestEnv5TwoMan:
    @pytest.fixture(autouse=True, scope='class')
    def configure_env(self):
        assert load_env(__file__)

    @pytest.fixture(autouse=True, scope='class')
    def test_config(self):
        return TestConfig.getDefaultTestConfig()

    def test_kssh_with_two_man(self, test_config):
        with autoapprover(), outputs_audit_log(test_config, filename="/shared/ca.log", expected_number=2, additional_regexes={f"Two-man SignatureRequest id=.* approved by ": 2}):
            assert_contains_hash(test_config.expected_hash, run_command_with_agent(f"bin/kssh --request-realm {test_config.subteam}.ssh.root_everywhere -q -o StrictHostKeyChecking=no root@sshd-prod 'sha1sum /etc/unique'"))
            assert_contains_hash(test_config.expected_hash, run_command_with_agent(f"bin/kssh --request-realm {test_config.subteam}.ssh.root_everywhere -q -o StrictHostKeyChecking=no root@sshd-staging 'sha1sum /etc/unique'"))

    def test_kssh_with_two_man_no_approval(self, test_config):
        with outputs_audit_log(test_config, filename="/shared/ca.log", expected_number=0):
            with pytest.raises(subprocess.CalledProcessError):
                run_command_with_agent(f"bin/kssh --request-realm {test_config.subteam}.ssh.root_everywhere -q -o StrictHostKeyChecking=no root@sshd-prod 'sha1sum /etc/unique'")

    def test_kssh_without_requested_realm(self, test_config):
        with outputs_audit_log(test_config, filename="/shared/ca.log", expected_number=2):
            with pytest.raises(subprocess.CalledProcessError):
                run_command_with_agent(f"bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod 'sha1sum /etc/unique'")
            with pytest.raises(subprocess.CalledProcessError):
                run_command_with_agent(f"bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod 'sha1sum /etc/unique'")
