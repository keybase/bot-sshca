#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

export "KEYBASE_USERNAME=$BOT_USERNAME"
export "KEYBASE_PAPERKEY=$BOT_PAPERKEY"

# For some reason it is necessary to touch a file in /shared/ in order to get the volume permissions to work correctly
# when keybaseca generate runs
touch /shared/.keep

# Generate the env files that will be used for tests
mkdir -p tests/generated-env
ls tests/envFiles/ | xargs -I {} -- bash -c 'cat tests/envFiles/{} | envsubst > tests/generated-env/{}'

nohup bash -c "KEYBASE_RUN_MODE=prod kbfsfuse /keybase | grep -v 'ERROR Mounting the filesystem failed' &"
sleep 3
keybase oneshot
touch /shared/ready
python3 tests/bot-entrypoint.py
