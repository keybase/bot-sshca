#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# For some reason it is necessary to touch a file in /mnt/ in order to get the volume permissions to work correctly
# when keybaseca generate runs
touch /mnt/.keep

nohup bash -c "run_keybase -g &"
sleep 3
keybase oneshot --username $KEYBASE_USERNAME --paperkey "$KEYBASE_PAPERKEY"
touch /mnt/ready
python3 tests/bot-entrypoint.py
