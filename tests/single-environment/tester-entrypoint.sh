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
    keybase fs rm /keybase/team/$SUBTEAM/kssh-client.config || true
    keybase fs rm /keybase/team/$SUBTEAM_SECONDARY/kssh-client.config || true
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

# Test 1: kssh works
bin/kssh -q -o StrictHostKeyChecking=no root@sshd "echo 'kssh passed test 1: It works!'"
# Test 2: kssh reuses keys. Checked by making sure it finishes quickly
timeout 0.5 bin/kssh -q -o StrictHostKeyChecking=no root@sshd "echo 'kssh passed test 2: Reuse unexpired keys'"
# Test 3: kssh generates a new key once the key is expired. Test this by sticking an expired key in the file and
# checking that it connects
clear_keys && mv tests/testFiles/expired ~/.ssh/keybase-signed-key-- && mv tests/testFiles/expired.pub ~/.ssh/keybase-signed-key--.pub && mv tests/testFiles/expired-cert.pub ~/.ssh/keybase-signed-key---cert.pub
bin/kssh -q -o StrictHostKeyChecking=no root@sshd "echo 'kssh passed test 3: Renew expired keys'"

# The next series of tests are about what happens when there are multiple teams configured. So we create a kssh-client.config
# file in $SUBTEAM_SECONDARY in order to make kssh think that there are two active teams with CA bots.
keybase fs cp /keybase/team/$SUBTEAM/kssh-client.config /keybase/team/$SUBTEAM_SECONDARY/kssh-client.config
! (clear_keys && bin/kssh -o StrictHostKeyChecking=no root@sshd) > /dev/null
clear_keys
OUTPUT=`bin/kssh -o StrictHostKeyChecking=no root@sshd || true`
if [[ $OUTPUT == *"Found 2 config files"* ]]; then
    echo 'kssh passed test 4: Rejects multiple teams'
else
    exit 1
fi
clear_keys && bin/kssh --team $SUBTEAM -o StrictHostKeyChecking=no root@sshd "echo 'kssh passed test 5: Works with specified --team flag'"
clear_keys && bin/kssh --set-default-team $SUBTEAM
clear_keys && bin/kssh -o StrictHostKeyChecking=no root@sshd "echo 'kssh passed test 6: Uses the default team'"
clear_keys && bin/kssh --set-default-team $SUBTEAM_SECONDARY
clear_keys && bin/kssh --team $SUBTEAM -o StrictHostKeyChecking=no root@sshd "echo 'kssh passed test 7: --team overrides the default team'"

# This tests the audit log feature
cat /mnt/ca.log | python3 ~/tests/integrationTestUtils.py logcheck 5 "root"
echo "kssh passed test 8: ca bot produces correct audit logs"

cleanup