import json
import subprocess

import pytest

from lib import assert_contains_hash, clear_keys, load_env, outputs_audit_log, run_command, SUBTEAM, BOT_USERNAME

class TestEnv4:
    @pytest.fixture(autouse=True, scope='class')
    def configure_env(self):
        load_env(__file__)

    @outputs_audit_log(filename="/mnt/ca.log", expected_number=0)
    def test_kssh_no_config_files(self):
        # Test that it can't find any config files
        for s in ['user@sshd-staging', 'root@sshd-staging', 'user@sshd-prod', 'root@sshd-prod']:
            try:
                run_command("""bin/kssh -q -o StrictHostKeyChecking=no %s "sha1sum /etc/unique" """ % s)
                assert False
            except subprocess.CalledProcessError as e:
                assert b"Did not find any config files in KBFS" in e.output

    def test_kssh_spoofed_config(self):
        # Test that even when kssh is forced to run by a spoofed config, the CA bot ignores messages that are in the
        # wrong channel
        client_config = json.dumps({'teamname': f"{SUBTEAM}.ssh", "channelname": "", "botname": BOT_USERNAME})
        run_command(f"echo '{client_config}' | keybase fs write /keybase/team/{SUBTEAM}.ssh/kssh-client.config")
        for s in ['user@sshd-staging', 'root@sshd-staging', 'user@sshd-prod', 'root@sshd-prod']:
            try:
                run_command("""bin/kssh -q -o StrictHostKeyChecking=no %s "sha1sum /etc/unique" """ % s)
                assert False
            except subprocess.CalledProcessError as e:
                assert b"Failed to get a signed key from the CA: timed out while waiting for a response from the CA" in e.output
