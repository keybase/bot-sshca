#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

if [ -e "/mnt/keybase-ca-key" ]
then
  nohup bash -c "KEYBASE_RUN_MODE=prod kbfsfuse /keybase | grep -v 'ERROR Mounting the filesystem failed' &"
  sleep 3
  keybase oneshot
  /home/keybase/bin/keybaseca service
else
  echo "keybase-ca-key file not found. Exiting."
fi
