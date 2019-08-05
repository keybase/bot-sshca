import json
import subprocess

import pytest

from lib import UtilitiesLib
from lib import SUBTEAM, SUBTEAM_SECONDARY, USERNAME, BOT_USERNAME, EXPECTED_HASH

class TestEnv4UserNotInConfiguredTeams:
    @pytest.fixture(autouse=True, scope='class')
    def configure_env(self, test_lib):
        assert test_lib.load_env(__file__)

    @pytest.fixture(autouse=True, scope='class')
    def test_lib(self):
        return UtilitiesLib(SUBTEAM, SUBTEAM_SECONDARY, USERNAME, BOT_USERNAME, EXPECTED_HASH)

    def test_kssh_no_config_files(self, test_lib):
        # Test that it can't find any config files
        with test_lib.outputs_audit_log(filename="/mnt/ca.log", expected_number=0):
            for s in ['user@sshd-staging', 'root@sshd-staging', 'user@sshd-prod', 'root@sshd-prod']:
                try:
                    test_lib.run_command(f"""bin/kssh -q -o StrictHostKeyChecking=no {s} "sha1sum /etc/unique" """)
                    assert False
                except subprocess.CalledProcessError as e:
                    assert b"Did not find any config files in KBFS" in e.output

    def test_kssh_spoofed_config(self, test_lib):
        # Test that even when kssh is forced to run by a spoofed config, the CA bot ignores messages that are in the
        # wrong channel
        with test_lib.outputs_audit_log(filename="/mnt/ca.log", expected_number=0):
            client_config = json.dumps({'teamname': f"{test_lib.subteam}.ssh", "channelname": "", "botname": BOT_USERNAME})
            test_lib.run_command(f"echo '{client_config}' | keybase fs write /keybase/team/{test_lib.subteam}.ssh/kssh-client.config")
            for s in ['user@sshd-staging', 'root@sshd-staging', 'user@sshd-prod', 'root@sshd-prod']:
                try:
                    test_lib.run_command(f"""bin/kssh -q -o StrictHostKeyChecking=no {s} "sha1sum /etc/unique" """)
                    assert False
                except subprocess.CalledProcessError as e:
                    assert b"Failed to get a signed key from the CA: timed out while waiting for a response from the CA" in e.output
