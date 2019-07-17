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
bin/keybaseca -c tests/single-environment/keybaseca.config generate --overwrite-existing-key
bin/keybaseca -c tests/single-environment/keybaseca.config service