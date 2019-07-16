#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# For some reason it is necessary to touch a file in /mnt/ in order to get the volume permissions to work correctly
# when keybaseca generate runs
touch /mnt/.keep

nohup bash -c "run_keybase -g &"
sleep 3
keybase oneshot --username $KEYBASE_USERNAME --paperkey "$PAPERKEY"
bin/keybaseca --wipe-all-configs
bin/keybaseca --wipe-logs -c tests/simple/keybaseca.config || true
bin/keybaseca -c tests/simple/keybaseca.config generate --overwrite-existing-key
bin/keybaseca -c tests/simple/keybaseca.config service