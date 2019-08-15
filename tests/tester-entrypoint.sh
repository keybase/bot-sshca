#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

nohup bash -c "run_keybase -g &"

# Sleep until the CA bot has started
while ! [ -f /shared/ready ];
do
    sleep 1
done
echo ""
sleep 2

keybase oneshot --username $KSSH_USERNAME --paperkey "$KSSH_PAPERKEY"

echo "========================= Launched Keybase, starting tests... ========================="

pytest -x --verbose ~/tests/
