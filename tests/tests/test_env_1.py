import hashlib
import os
import subprocess
import time

import pytest

from lib import assert_contains_hash, EXPECTED_HASH, load_env, outputs_audit_log, run_command, simulate_two_teams, SUBTEAM, SUBTEAM_SECONDARY

test_env_1_log_filename = f"/keybase/team/{SUBTEAM}.ssh.staging/ca.log"
class TestEnv1:

    @pytest.fixture(autouse=True, scope='class')
    def configure_env(self):
        assert load_env(__file__)

    @outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    def test_kssh_staging_user(self):
        # Test ksshing into staging as user
        assert_contains_hash(run_command("""bin/kssh -q -o StrictHostKeyChecking=no user@sshd-staging "sha1sum /etc/unique" """))

    @outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    def test_kssh_staging_root(self):
        # Test ksshing into staging as user
        assert_contains_hash(run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-staging "sha1sum /etc/unique" """))

    @outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    def test_kssh_prod_root(self):
        # Test ksshing into prod as root
        assert_contains_hash(run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))

    @outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    def test_kssh_reject_prod_user(self):
        # Test that we can't kssh into prod as user since we aren't in the correct team for that
        try:
            run_command("""bin/kssh -o StrictHostKeyChecking=no user@sshd-prod "sha1sum /etc/unique" 2>&1 """)
            assert False
        except subprocess.CalledProcessError as e:
            assert b"Permission denied" in e.output
            assert EXPECTED_HASH not in e.output

    @outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    def test_kssh_reuse(self):
        # Test that kssh reuses unexpired keys
        assert_contains_hash(run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))
        start = time.time()
        assert_contains_hash(run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))
        elapsed = time.time() - start
        assert elapsed < 0.5

    @outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    def test_kssh_regenerate_expired_keys(self):
        # Test that kssh reprovisions a key when the stored keys are expired
        run_command("mv ~/tests/testFiles/expired ~/.ssh/keybase-signed-key-- && mv ~/tests/testFiles/expired.pub ~/.ssh/keybase-signed-key--.pub && mv ~/tests/testFiles/expired-cert.pub ~/.ssh/keybase-signed-key---cert.pub")
        assert_contains_hash(run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))

    @outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    def test_kssh_provision(self):
        # Test the `kssh --provision` flag
        # we have to run all of the below commands in one run_command call so that environment variables are shared
        # so ssh-agent can work
        output = run_command("""
        eval `ssh-agent -s`
        bin/kssh --provision
        ssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique"
        echo -n foo > /tmp/foo
        scp /tmp/foo root@sshd-prod:/tmp/foo
        ssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /tmp/foo"
        """)
        assert_contains_hash(output)
        assert hashlib.sha1(b"foo").hexdigest().encode('utf-8') in output

    @outputs_audit_log(filename=test_env_1_log_filename, expected_number=0)
    @simulate_two_teams
    def test_kssh_errors_on_two_teams(self):
        # Test that kssh does not run if there are multiple teams, no client config, and no --team flag
        try:
            run_command("bin/kssh root@sshd-prod")
            assert False
        except subprocess.CalledProcessError as e:
            assert b"Found 2 config files" in e.output

    @outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    @simulate_two_teams
    def test_kssh_team_flag(self):
        # Test that kssh works with the --team flag
        assert_contains_hash(run_command(f"bin/kssh --team {SUBTEAM}.ssh -q -o StrictHostKeyChecking=no root@sshd-prod 'sha1sum /etc/unique'"))

    @outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    @simulate_two_teams
    def test_kssh_set_default_team(self):
        # Test that kssh works with the --set-default-team flag
        run_command(f"bin/kssh --set-default-team {SUBTEAM}.ssh")
        assert_contains_hash(run_command("bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod 'sha1sum /etc/unique'"))

    @outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    @simulate_two_teams
    def test_kssh_override_default_team(self):
        # Test that the --team flag overrides the local config file
        run_command(f"bin/kssh --set-default-team {SUBTEAM_SECONDARY}")
        assert_contains_hash(run_command(f"bin/kssh --team {SUBTEAM}.ssh -q -o StrictHostKeyChecking=no root@sshd-prod 'sha1sum /etc/unique'"))

    def test_keybaseca_backup(self):
        # Test the keybaseca backup command by reading and verifying the private key stored in /mnt/cakey.backup
        run_command("mkdir -p /tmp/ssh/")
        run_command("chown -R keybase:keybase /tmp/ssh/")
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
        run_command("chmod 0600 /tmp/ssh/cakey")
        output = run_command("ssh-keygen -y -e -f /tmp/ssh/cakey")
        assert b"BEGIN SSH2 PUBLIC KEY" in output
