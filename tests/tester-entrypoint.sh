#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

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

pytest -x --verbose ~/tests/
