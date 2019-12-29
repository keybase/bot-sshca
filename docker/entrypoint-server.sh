#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

if [ -e "/mnt/keybase-ca-key" ]
then
  nohup bash -c "run_keybase -g 2>&1 | grep -v 'KBFS failed to FUSE mount' &"
  sleep 3
  keybase oneshot
  /home/keybase/bin/keybaseca service
fi
