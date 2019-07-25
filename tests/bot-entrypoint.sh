#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# For some reason it is necessary to touch a file in /mnt/ in order to get the volume permissions to work correctly
# when keybaseca generate runs
touch /mnt/.keep

nohup bash -c "run_keybase -g &"
sleep 3

source tests/generated-env/env-1-simple-tests
keybase oneshot --username $KEYBASE_USERNAME --paperkey "$KEYBASE_PAPERKEY"
bin/keybaseca --wipe-all-configs
bin/keybaseca --wipe-logs || true
bin/keybaseca generate --overwrite-existing-key
echo yes | bin/keybaseca backup > /mnt/cakey.backup
bin/keybaseca service
