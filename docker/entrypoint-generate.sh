#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

nohup bash -c "run_keybase -g 2>&1 | grep -v 'KBFS failed to FUSE mount' &"
sleep 3
keybase oneshot
bin/keybaseca generate
