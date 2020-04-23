import hashlib
import subprocess
import time

import pytest

from lib import TestConfig, load_env, outputs_audit_log, assert_contains_hash, run_command, simulate_two_teams, get_principals, run_command_with_agent

test_env_1_log_filename = f"/keybase/team/{TestConfig.getDefaultTestConfig().subteam}.ssh.staging/ca.log"
class TestMultiTeamStrictLogging:
    @pytest.fixture(autouse=True, scope='class')
    def configure_env(self):
        assert load_env(__file__)

    @pytest.fixture(autouse=True, scope='class')
    def test_config(self):
        return TestConfig.getDefaultTestConfig()

    def test_ping_pong_command(self, test_config):
        run_command(f"keybase chat send --channel ssh-provision {test_config.subteam}.ssh 'ping @{test_config.bot_username}'")
        time.sleep(5)
        recent_messages = run_command(f"keybase chat list-unread --since 1m")
        print(recent_messages)
        assert (b"pong @%s" % test_config.username.encode('utf-8')) in recent_messages

    def test_kssh_staging_user(self, test_config):
        # Test ksshing into staging as user
        with outputs_audit_log(test_config, filename=test_env_1_log_filename, expected_number=1):
            assert_contains_hash(test_config.expected_hash, run_command_with_agent("""bin/kssh -q -o StrictHostKeyChecking=no user@sshd-staging "sha1sum /etc/unique" """))

    def test_kssh_staging_root(self, test_config):
        # Test ksshing into staging as user
        with outputs_audit_log(test_config, filename=test_env_1_log_filename, expected_number=1):
            assert_contains_hash(test_config.expected_hash, run_command_with_agent("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-staging "sha1sum /etc/unique" """))

    def test_kssh_prod_root(self, test_config):
        # Test ksshing into prod as root
        with outputs_audit_log(test_config, filename=test_env_1_log_filename, expected_number=1):
            assert_contains_hash(test_config.expected_hash, run_command_with_agent("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))

    def test_kssh_reject_prod_user(self, test_config):
        # Test that we can't kssh into prod as user since we aren't in the correct team for that
        with outputs_audit_log(test_config, filename=test_env_1_log_filename, expected_number=1):
            try:
                run_command_with_agent("""bin/kssh -o StrictHostKeyChecking=no user@sshd-prod "sha1sum /etc/unique" 2>&1 """)
                assert False
            except subprocess.CalledProcessError as e:
                assert b"Permission denied" in e.output
                assert test_config.expected_hash not in e.output

    def test_kssh_reuse(self, test_config):
        # Test that kssh reuses unexpired keys
        with outputs_audit_log(test_config, filename=test_env_1_log_filename, expected_number=1):
            assert_contains_hash(test_config.expected_hash, run_command_with_agent("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))
            start = time.time()
            assert_contains_hash(test_config.expected_hash, run_command_with_agent("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))
            elapsed = time.time() - start
            assert elapsed < 0.5

    def test_kssh_regenerate_expired_keys(self, test_config):
        # Test that kssh reprovisions a key when the stored keys are expired
        with outputs_audit_log(test_config, filename=test_env_1_log_filename, expected_number=1):
            run_command_with_agent("mv ~/tests/testFiles/expired ~/.ssh/keybase-signed-key-- && mv ~/tests/testFiles/expired.pub ~/.ssh/keybase-signed-key--.pub && mv ~/tests/testFiles/expired-cert.pub ~/.ssh/keybase-signed-key---cert.pub")
            assert_contains_hash(test_config.expected_hash, run_command_with_agent("""bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique" """))

    def test_kssh_provision(self, test_config):
        # Test the `kssh --provision` flag
        # we have to run all of the below commands in one run_command call so that environment variables are shared
        # so ssh-agent can work
        with outputs_audit_log(test_config, filename=test_env_1_log_filename, expected_number=1):
            output = run_command_with_agent("""
            bin/kssh --provision
            ssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /etc/unique"
            echo -n foo > /tmp/foo
            scp /tmp/foo root@sshd-prod:/tmp/foo
            ssh -q -o StrictHostKeyChecking=no root@sshd-prod "sha1sum /tmp/foo"
            """)
            assert_contains_hash(test_config.expected_hash, output)
            assert hashlib.sha1(b"foo").hexdigest().encode('utf-8') in output
        assert get_principals("~/.ssh/keybase-signed-key---cert.pub") == set([test_config.subteam + ".ssh.staging", test_config.subteam + ".ssh.root_everywhere"])

    def test_kssh_errors_on_two_bots(self, test_config):
        # Test that kssh does not run if there are multiple bots, no kssh config, and no --bot flag
        with simulate_two_teams(test_config), outputs_audit_log(test_config, filename=test_env_1_log_filename, expected_number=0):
            try:
                run_command_with_agent("bin/kssh root@sshd-prod")
                assert False
            except subprocess.CalledProcessError as e:
                assert b"Found 2 config files" in e.output

    def test_kssh_bot_flag(self, test_config):
        # Test that kssh works with the --bot flag
        with simulate_two_teams(test_config), outputs_audit_log(test_config, filename=test_env_1_log_filename, expected_number=1):
            assert_contains_hash(test_config.expected_hash, run_command_with_agent(f"bin/kssh --bot {test_config.bot_username} -q -o StrictHostKeyChecking=no root@sshd-prod 'sha1sum /etc/unique'"))

    def test_kssh_set_default_bot(self, test_config):
        # Test that kssh works with the --set-default-bot flag
        with simulate_two_teams(test_config), outputs_audit_log(test_config, filename=test_env_1_log_filename, expected_number=1):
            run_command_with_agent(f"bin/kssh --set-default-bot {test_config.bot_username}")
            assert_contains_hash(test_config.expected_hash, run_command_with_agent("bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod 'sha1sum /etc/unique'"))

    def test_kssh_override_default_bot(self, test_config):
        # Test that the --bot flag overrides the local config file
        with simulate_two_teams(test_config), outputs_audit_log(test_config, filename=test_env_1_log_filename, expected_number=1):
            run_command_with_agent(f"bin/kssh --set-default-bot otherbotname")
            assert_contains_hash(test_config.expected_hash, run_command_with_agent(f"bin/kssh --bot {test_config.bot_username} -q -o StrictHostKeyChecking=no root@sshd-prod 'sha1sum /etc/unique'"))

    def test_kssh_clear_default_bot(self, test_config):
        # Test that kssh --clear-default-bot clears the default bot
        with simulate_two_teams(test_config), outputs_audit_log(test_config, filename=test_env_1_log_filename, expected_number=0):
            run_command_with_agent(f"bin/kssh --set-default-bot {test_config.bot_username}")
            run_command_with_agent("bin/kssh --clear-default-bot")
            try:
                # No default set and no bot specified so it will error out
                run_command_with_agent("bin/kssh root@sshd-prod")
                assert False
            except subprocess.CalledProcessError as e:
                assert b"Found 2 config files" in e.output

    def test_keybaseca_backup(self):
        # Test the keybaseca backup command by reading and verifying the private key stored in /shared/cakey.backup
        run_command("mkdir -p /tmp/ssh/")
        run_command("chown -R keybase:keybase /tmp/ssh/")
        with open('/shared/cakey.backup') as f:
            keyLines = []
            add = False
            for line in f:
                if "----" in line and "PRIVATE" in line and "BEGIN" in line:
                    add = True
                if add:
                    keyLines.append(line.strip())
                if "----" in line and "PRIVATE" in line and "END" in line:
                    add = False
        key = '\n'.join(keyLines) + '\n'
        with open('/tmp/ssh/cakey', 'w+') as f:
            f.write(key)
        run_command("chmod 0600 /tmp/ssh/cakey")
        output = run_command("ssh-keygen -y -e -f /tmp/ssh/cakey")
        assert b"BEGIN SSH2 PUBLIC KEY" in output

    def test_keybaseca_sign(self, test_config):
        # Stdout contains a useful message
        with open('/shared/keybaseca-sign.out') as f:
            assert "Provisioned new certificate" in f.read()

        # SSH with that certificate should just work for every team
        assert_contains_hash(test_config.expected_hash, run_command(f"ssh -q -o StrictHostKeyChecking=no -i /shared/userkey user@sshd-prod 'sha1sum /etc/unique'"))
        assert_contains_hash(test_config.expected_hash, run_command(f"ssh -q -o StrictHostKeyChecking=no -i /shared/userkey root@sshd-prod 'sha1sum /etc/unique'"))
        assert_contains_hash(test_config.expected_hash, run_command(f"ssh -q -o StrictHostKeyChecking=no -i /shared/userkey user@sshd-staging 'sha1sum /etc/unique'"))
        assert_contains_hash(test_config.expected_hash, run_command(f"ssh -q -o StrictHostKeyChecking=no -i /shared/userkey root@sshd-prod 'sha1sum /etc/unique'"))

        # Checking that it actually contains the correct principals
        assert get_principals("/shared/userkey-cert.pub") == set(test_config.subteams)

    def test_kssh_default_user(self, test_config):
        # Set the default user to root
        run_command_with_agent("bin/kssh --set-default-user root")
        # A normal SSH connection
        assert_contains_hash(test_config.expected_hash, run_command_with_agent("bin/kssh -q -o StrictHostKeyChecking=no sshd-prod 'sha1sum /etc/unique'"))
        assert b"root" in run_command_with_agent("bin/kssh -q -o StrictHostKeyChecking=no sshd-prod 'whoami'")
        # A proxy jump (relies on the ssh agent)
        assert_contains_hash(test_config.expected_hash, run_command_with_agent("bin/kssh -q -o StrictHostKeyChecking=no -J sshd-staging sshd-prod 'sha1sum /etc/unique'"))
        # Reset the default user
        run_command_with_agent("bin/kssh --clear-default-user")

    def test_kssh_alternate_binary(self, test_config):
        # Test it by creating another keybase binary earlier in the path and running kssh. This isn't a perfect test but it is
        # enough to smoketest it
        run_command("echo '#!/bin/bash' | sudo tee /usr/local/bin/keybase")
        run_command("sudo chmod +x /usr/local/bin/keybase")
        try:
            run_command("bin/kssh --set-keybase-binary /usr/bin/keybase")
            assert_contains_hash(test_config.expected_hash, run_command_with_agent("bin/kssh -q -o StrictHostKeyChecking=no user@sshd-staging 'sha1sum /etc/unique'"))
            run_command("bin/kssh --set-keybase-binary ''")
        finally:
            run_command("sudo rm /usr/local/bin/keybase")