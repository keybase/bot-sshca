#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# For some reason it is necessary to touch a file in /mnt/ in order to get the volume permissions to work correctly
# when keybaseca generate runs
touch /mnt/.keep

# Generate the env files that will be used for tests
source tests/env.sh
mkdir -p tests/generated-env
ls tests/envFiles/ | xargs -I {} -- bash -c 'cat tests/envFiles/{} | envsubst > tests/generated-env/{}'

nohup bash -c "run_keybase -g &"
sleep 3
keybase oneshot --username $BOT_USERNAME --paperkey "$BOT_PAPERKEY"
touch /mnt/ready
python3 tests/bot-entrypoint.py
