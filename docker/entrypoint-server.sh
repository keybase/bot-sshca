#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# chown as root
chown -R keybase:keybase /mnt

# Run everything else as the keybase user
sudo -i -u keybase bash << EOF
nohup bash -c "run_keybase -g -f &"
sleep 3
keybase oneshot --username $KEYBASE_USERNAME --paperkey "$KEYBASE_PAPERKEY"
source docker/env.sh && bin/keybaseca service
EOF
