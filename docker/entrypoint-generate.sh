#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# chown as root
chown -R keybase:keybase /mnt

# Run everything else as the keybase user
sudo -i -u keybase bash << EOF
export "FORCE_WRITE=$FORCE_WRITE"
nohup bash -c "KEYBASE_RUN_MODE=prod kbfsfuse /keybase | grep -v 'ERROR Mounting the filesystem failed' &"
sleep 3
keybase oneshot
bin/keybaseca generate
EOF
