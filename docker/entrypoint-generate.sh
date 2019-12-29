#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

nohup bash -c "KEYBASE_RUN_MODE=prod kbfsfuse /keybase | grep -v 'ERROR Mounting the filesystem failed' &"
sleep 3
keybase oneshot
bin/keybaseca generate
