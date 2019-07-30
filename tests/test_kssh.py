import hashlib
import os
import subprocess
import time

import pytest
import requests

import lib

@pytest.fixture(autouse=True)
def run_around_tests():
    lib.clear_keys()
    lib.clear_local_config()
    # Calling yield triggers the test execution
    yield

test_env_1_log_filename = "/keybase/team/%s.ssh.staging/ca.log" % os.environ['SUBTEAM']
class TestEnv1:

    @pytest.fixture(autouse=True, scope='class')
    def configure_env(self):
        assert requests.get("http://ca-bot:8080/load_env?filename=env-1-integration-tests").content == b"OK"

    @lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    def test_kssh_staging_user(self):
        # Test ksshing into staging as user
        lib.assert_contains_hash(lib.run_command("""bin/kssh -q -o StrictHostKeyChecking=no user@sshd-staging "sha1sum /etc/unique" """))

    @lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    def test_kssh_staging_root(self):
        # Test ksshing into staging as user
        lib.assert_contains_hash(lib.run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-staging "sha1sum /etc/unique" """))

    @lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    def test_kssh_prod_root(self):
        # Test ksshing into prod as root
        lib.assert_contains_hash(lib.run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))

    @lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    def test_kssh_reject_prod_user(self):
        # Test that we can't kssh into prod as user since we aren't in the correct team for that
        try:
            lib.run_command("""bin/kssh -o StrictHostKeyChecking=no user@sshd-prod "sha1sum /etc/unique" 2>&1 """)
            assert False
        except subprocess.CalledProcessError as e:
            assert b"Permission denied" in e.output
            assert lib.EXPECTED_HASH not in e.output

    @lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    def test_kssh_reuse(self):
        # Test that kssh reuses unexpired keys
        lib.assert_contains_hash(lib.run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))
        start = time.time()
        lib.assert_contains_hash(lib.run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))
        elapsed = time.time() - start
        assert elapsed < 0.5

    @lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    def test_kssh_regenerate_expired_keys(self):
        # Test that kssh reprovisions a key when the stored keys are expired
        lib.run_command("ls ~/")
        lib.run_command("mv ~/tests/testFiles/expired ~/.ssh/keybase-signed-key-- && mv ~/tests/testFiles/expired.pub ~/.ssh/keybase-signed-key--.pub && mv ~/tests/testFiles/expired-cert.pub ~/.ssh/keybase-signed-key---cert.pub")
        lib.assert_contains_hash(lib.run_command("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))

    @lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    def test_kssh_provision(self):
        # Test the `kssh --provision` flag
        # we have to run all of the below commands in one lib.run_command call so that environment variables are shared
        # so ssh-agent can work
        output = lib.run_command("""
        eval `ssh-agent -s`
        bin/kssh --provision
        ssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique"
        echo -n foo > /tmp/foo
        scp /tmp/foo root@sshd-prod:/tmp/foo
        ssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /tmp/foo"
        """)
        lib.assert_contains_hash(output)
        assert hashlib.sha1(b"foo").hexdigest().encode('utf-8') in output

    @lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=0)
    @lib.simulate_two_teams
    def test_kssh_errors_on_two_teams(self):
        # Test that kssh does not run if there are multiple teams, no client config, and no --team flag
        try:
            lib.run_command("bin/kssh root@sshd-prod")
            assert False
        except subprocess.CalledProcessError as e:
            assert b"Found 2 config files" in e.output

    @lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    @lib.simulate_two_teams
    def test_kssh_team_flag(self):
        # Test that kssh works with the --team flag
        lib.assert_contains_hash(lib.run_command("bin/kssh --team %s.ssh -q -o StrictHostKeyChecking=no root@sshd-prod 'sha1sum /etc/unique'" % os.environ['SUBTEAM']))

    @lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    @lib.simulate_two_teams
    def test_kssh_set_default_team(self):
        # Test that kssh works with the --set-default-team flag
        lib.run_command("bin/kssh --set-default-team %s.ssh" % os.environ['SUBTEAM'])
        lib.assert_contains_hash(lib.run_command("bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod 'sha1sum /etc/unique'"))

    @lib.outputs_audit_log(filename=test_env_1_log_filename, expected_number=1)
    @lib.simulate_two_teams
    def test_kssh_override_default_team(self):
        # Test that the --team flag overrides the local config file
        lib.run_command("bin/kssh --set-default-team %s" % os.environ['SUBTEAM_SECONDARY'])
        lib.assert_contains_hash(lib.run_command("bin/kssh --team %s.ssh -q -o StrictHostKeyChecking=no root@sshd-prod 'sha1sum /etc/unique'" % os.environ['SUBTEAM']))

    def test_keybaseca_backup(self):
        # Test the keybaseca backup command by reading and verifying the private key stored in /mnt/cakey.backup
        lib.run_command("mkdir -p /tmp/ssh/")
        lib.run_command("chown -R keybase:keybase /tmp/ssh/")
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
        lib.run_command("chmod 0600 /tmp/ssh/cakey")
        output = lib.run_command("ssh-keygen -y -e -f /tmp/ssh/cakey")
        assert b"BEGIN SSH2 PUBLIC KEY" in output

class TestEnv2:
    @pytest.fixture(autouse=True, scope='class')
    def configure_env(self):
        assert requests.get("http://ca-bot:8080/load_env?filename=env-2-log-to-fs").content == b"OK"

    @lib.outputs_audit_log(filename="/mnt/ca.log", expected_number=1)
    def test_kssh(self):
        # Test ksshing into staging as user
        lib.assert_contains_hash(lib.run_command("""bin/kssh -q -o StrictHostKeyChecking=no user@sshd-staging "sha1sum /etc/unique" """))


def pytest_sessionfinish(session, exitstatus):
    # Automatically run after all tests in order to ensure that no kssh-client config files stick around
    lib.run_command("keybase fs rm /keybase/team/%s.ssh/kssh-client.config || true" % os.environ['SUBTEAM'])
    lib.run_command("keybase fs rm /keybase/team/%s.ssh.staging/kssh-client.config || true" % os.environ['SUBTEAM'])
    lib.run_command("keybase fs rm /keybase/team/%s.ssh.prod/kssh-client.config || true" % os.environ['SUBTEAM'])
    lib.run_command("keybase fs rm /keybase/team/%s.ssh.root_everywhere/kssh-client.config || true" % os.environ['SUBTEAM'])
    lib.run_command("keybase fs rm /keybase/team/%s/kssh-client.config || true" % os.environ['SUBTEAM_SECONDARY'])
