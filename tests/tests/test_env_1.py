import hashlib
import os
import subprocess
import time

import pytest

from lib import UtilitiesLib
from lib import SUBTEAM, SUBTEAM_SECONDARY, USERNAME, BOT_USERNAME, EXPECTED_HASH

test_env_1_log_filename = f"/keybase/team/{SUBTEAM}.ssh.staging/ca.log"
class TestMultiTeamStrictLogging:
    @pytest.fixture(autouse=True, scope='class')
    def configure_env(self, test_lib):
        assert test_lib.load_env(__file__)

    @pytest.fixture(autouse=True, scope='class')
    def test_lib(self):
        return UtilitiesLib(SUBTEAM, SUBTEAM_SECONDARY, USERNAME, BOT_USERNAME, EXPECTED_HASH)

    def test_kssh_staging_user(self, test_lib):
        # Test ksshing into staging as user
        with test_lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1):
            test_lib.assert_contains_hash(test_lib.run_command("""bin/kssh -q -o StrictHostKeyChecking=no user@sshd-staging "sha1sum /etc/unique" """))

    def test_kssh_staging_root(self, test_lib):
        # Test ksshing into staging as user
        with test_lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1):
            test_lib.assert_contains_hash(test_lib.run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-staging "sha1sum /etc/unique" """))

    def test_kssh_prod_root(self, test_lib):
        # Test ksshing into prod as root
        with test_lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1):
            test_lib.assert_contains_hash(test_lib.run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))

    def test_kssh_reject_prod_user(self, test_lib):
        # Test that we can't kssh into prod as user since we aren't in the correct team for that
        with test_lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1):
            try:
                test_lib.run_command("""bin/kssh -o StrictHostKeyChecking=no user@sshd-prod "sha1sum /etc/unique" 2>&1 """)
                assert False
            except subprocess.CalledProcessError as e:
                assert b"Permission denied" in e.output
                assert EXPECTED_HASH not in e.output

    def test_kssh_reuse(self, test_lib):
        # Test that kssh reuses unexpired keys
        with test_lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1):
            test_lib.assert_contains_hash(test_lib.run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))
            start = time.time()
            test_lib.assert_contains_hash(test_lib.run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))
            elapsed = time.time() - start
            assert elapsed < 0.5

    def test_kssh_regenerate_expired_keys(self, test_lib):
        # Test that kssh reprovisions a key when the stored keys are expired
        with test_lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1):
            test_lib.run_command("mv ~/tests/testFiles/expired ~/.ssh/keybase-signed-key-- && mv ~/tests/testFiles/expired.pub ~/.ssh/keybase-signed-key--.pub && mv ~/tests/testFiles/expired-cert.pub ~/.ssh/keybase-signed-key---cert.pub")
            test_lib.assert_contains_hash(test_lib.run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))

    def test_kssh_provision(self, test_lib):
        # Test the `kssh --provision` flag
        # we have to run all of the below commands in one run_command call so that environment variables are shared
        # so ssh-agent can work
        with test_lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1):
            output = test_lib.run_command("""
            eval `ssh-agent -s`
            bin/kssh --provision
            ssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique"
            echo -n foo > /tmp/foo
            scp /tmp/foo root@sshd-prod:/tmp/foo
            ssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /tmp/foo"
            """)
            test_lib.assert_contains_hash(output)
            assert hashlib.sha1(b"foo").hexdigest().encode('utf-8') in output

    def test_kssh_errors_on_two_bots(self, test_lib):
        # Test that kssh does not run if there are multiple bots, no kssh config, and no --bot flag
        with test_lib.simulate_two_teams(), test_lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=0):
            try:
                test_lib.run_command("bin/kssh root@sshd-prod")
                assert False
            except subprocess.CalledProcessError as e:
                assert b"Found 2 config files" in e.output

    def test_kssh_bot_flag(self, test_lib):
        # Test that kssh works with the --bot flag
        with test_lib.simulate_two_teams(), test_lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1):
            test_lib.assert_contains_hash(test_lib.run_command(f"bin/kssh --bot {test_lib.bot_username} -q -o StrictHostKeyChecking=no root@sshd-prod 'sha1sum /etc/unique'"))

    def test_kssh_set_default_bot(self, test_lib):
        # Test that kssh works with the --set-default-bot flag
        with test_lib.simulate_two_teams(), test_lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1):
            test_lib.run_command(f"bin/kssh --set-default-bot {test_lib.bot_username}")
            test_lib.assert_contains_hash(test_lib.run_command("bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod 'sha1sum /etc/unique'"))

    def test_kssh_override_default_bot(self, test_lib):
        # Test that the --bot flag overrides the local config file
        with test_lib.simulate_two_teams(), test_lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1):
            test_lib.run_command(f"bin/kssh --set-default-bot otherbotname")
            test_lib.assert_contains_hash(test_lib.run_command(f"bin/kssh --bot {test_lib.bot_username} -q -o StrictHostKeyChecking=no root@sshd-prod 'sha1sum /etc/unique'"))

    def test_kssh_clear_default_bot(self, test_lib):
        # Test that kssh --clear-default-bot clears the default bot
        with test_lib.simulate_two_teams(), test_lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=0):
            test_lib.run_command(f"bin/kssh --set-default-bot {test_lib.bot_username}")
            test_lib.run_command("bin/kssh --clear-default-bot")
            try:
                # No default set and no bot specified so it will error out
                test_lib.run_command("bin/kssh root@sshd-prod")
                assert False
            except subprocess.CalledProcessError as e:
                assert b"Found 2 config files" in e.output

    def test_keybaseca_backup(self, test_lib):
        # Test the keybaseca backup command by reading and verifying the private key stored in /mnt/cakey.backup
        test_lib.run_command("mkdir -p /tmp/ssh/")
        test_lib.run_command("chown -R keybase:keybase /tmp/ssh/")
        with open('/mnt/cakey.backup') as f:
            keyLines = []
            add = False
            for line in f:
                if "----" in line and "PRIVATE" in line and "BEGIN" in line:
                    add = True
                if add:
                    keyLines.append(line)
                if "----" in line and "PRIVATE" in line and "END" in line:
                    add = False
        key = '\n'.join(keyLines)
        print(key)
        with open('/tmp/ssh/cakey', 'w+') as f:
            f.write(key)
        test_lib.run_command("chmod 0600 /tmp/ssh/cakey")
        output = test_lib.run_command("ssh-keygen -y -e -f /tmp/ssh/cakey")
        assert b"BEGIN SSH2 PUBLIC KEY" in output
