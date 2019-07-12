#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# chown as root
chown keybase:keybase /mnt

# Run everything else as the keybase user
sudo -i -u keybase bash << EOF
nohup bash -c "run_keybase -g &"
sleep 3
keybase oneshot --username $KEYBASE_USERNAME --paperkey "$PAPERKEY"
bin/keybaseca -c /mnt/keybaseca.config generate
EOF