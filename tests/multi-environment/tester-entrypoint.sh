#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# This file contains all of the kssh tests. Tests are simply written in bash and a failed test is signalled to the
# test runner via a non-zero exit code for this script.

# If something crashes, report what line number in order to make it easier to debug
err_report() {
    echo "Error on line $1"
    cleanup
}
trap 'err_report $LINENO' ERR

# Ensure that we always run cleanup on test exit. Note that the cleanup function should be idempotent since it may be
# run multiple times
cleanup() {
    keybase fs rm /keybase/team/$SUBTEAM.ssh.staging/kssh-client.config || true
    keybase fs rm /keybase/team/$SUBTEAM.ssh.prod/kssh-client.config || true
    keybase fs rm /keybase/team/$SUBTEAM.ssh.root_everywhere/kssh-client.config || true
}
trap cleanup EXIT

# Delete cached signed keys
clear_keys() {
    rm -rf ~/.ssh/keybase-signed-key*
}

nohup bash -c "run_keybase -g &"

# Sleep until the CA bot has started
while ! [ -f /mnt/keybase-ca-key.pub ];
do
    sleep 1
done
echo ""
sleep 2

keybase oneshot --username $KEYBASE_USERNAME --paperkey "$PAPERKEY"

echo "========================= Launched Keybase, starting tests... ========================="

# Tests that show it is putting the correct principals in the keys
clear_keys && bin/kssh -q -o StrictHostKeyChecking=no user@sshd-staging "echo 'kssh passed test 1: SSH into staging as user!'"
clear_keys && bin/kssh -q -o StrictHostKeyChecking=no root@sshd-staging "echo 'kssh passed test 2: SSH into staging as root!'"
clear_keys && bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "echo 'kssh passed test 3: SSH into prod as root!'"
clear_keys
!(bin/kssh -q -o StrictHostKeyChecking=no user@sshd-prod 2>&1 > /dev/null)
OUTPUT=`bin/kssh -o StrictHostKeyChecking=no user@sshd-prod 2>&1 || true`
if [[ $OUTPUT == *"Permission denied"* ]]; then
    echo 'kssh passed test 4: Does not sign the key for teams the user is not in'
else
    exit 1
fi

# Test 5: kssh reuses keys. Checked by making sure it finishes quickly
clear_keys && bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "echo -e ''"
timeout 0.5 bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "echo 'kssh passed test 5: Reuse unexpired keys'"
# Test 5: kssh generates a new key once the key is expired. Test this by sticking an expired key in the file and
# checking that it connects
clear_keys && mv tests/testFiles/expired ~/.ssh/keybase-signed-key-- && mv tests/testFiles/expired.pub ~/.ssh/keybase-signed-key--.pub && mv tests/testFiles/expired-cert.pub ~/.ssh/keybase-signed-key---cert.pub
bin/kssh -q -o StrictHostKeyChecking=no root@sshd-prod "echo 'kssh passed test 6: Renew expired keys'"

# The next series of tests are about what happens when there are multiple teams configured. So we create a kssh-client.config
# file in $SUBTEAM_SECONDARY in order to make kssh think that there are two active teams with CA bots.
keybase fs cp /keybase/team/$SUBTEAM.ssh.staging/kssh-client.config /keybase/team/$SUBTEAM_SECONDARY/kssh-client.config
! (clear_keys && bin/kssh -o StrictHostKeyChecking=no root@sshd-prod) > /dev/null
clear_keys
OUTPUT=`bin/kssh -o StrictHostKeyChecking=no root@sshd-prod || true`
if [[ $OUTPUT == *"Found 2 config files"* ]]; then
    echo 'kssh passed test 7: Rejects multiple teams'
else
    exit 1
fi
clear_keys && bin/kssh --team $SUBTEAM.ssh.staging -o StrictHostKeyChecking=no root@sshd-prod "echo 'kssh passed test 8: Works with specified --team flag'"
clear_keys && bin/kssh --set-default-team $SUBTEAM.ssh.staging
clear_keys && bin/kssh -o StrictHostKeyChecking=no root@sshd-prod "echo 'kssh passed test 9: Uses the default team'"
clear_keys && bin/kssh --set-default-team $SUBTEAM_SECONDARY.ssh.staging
clear_keys && bin/kssh --team $SUBTEAM.ssh.staging -o StrictHostKeyChecking=no root@sshd-prod "echo 'kssh passed test 10: --team overrides the default team'"

# This tests the audit log feature
sleep 2 # sleep to make sure ca-bot has synced all kbfs changes
keybase fs read /keybase/team/$SUBTEAM.ssh.staging/ca.log | python3 ~/tests/integrationTestUtils.py logcheck 9 "staging,root_everywhere"
echo "kssh passed test 11: ca bot produces correct audit logs"

cleanup