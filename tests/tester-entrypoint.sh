#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

nohup bash -c "run_keybase -g &"

sleep 10

keybase oneshot --username $KSSH_USERNAME --paperkey "$KSSH_PAPERKEY"

echo "========================= Launched Keybase, starting tests... ========================="

pytest -x --verbose ~/tests/
