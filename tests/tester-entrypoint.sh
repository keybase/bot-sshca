#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

export "KEYBASE_USERNAME=$KSSH_USERNAME"
export "KEYBASE_PAPERKEY=$KSSH_PAPERKEY"

nohup bash -c "run_keybase -g &"

# Sleep until the CA bot has started
while ! [ -f /shared/ready ];
do
    sleep 1
done
echo ""
sleep 2

keybase oneshot

echo "========================= Launched Keybase, starting tests... ========================="

pytest -x --verbose ~/tests/
