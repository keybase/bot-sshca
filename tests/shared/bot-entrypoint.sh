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
bin/keybaseca --wipe-logs -c $1 || true
bin/keybaseca -c $1 generate --overwrite-existing-key
echo yes | bin/keybaseca -c $1 backup > /mnt/cakey.backup
bin/keybaseca -c $1 service
