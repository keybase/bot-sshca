#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# chown as root
chown -R keybase:keybase /mnt

# Run everything else as the keybase user
sudo -i -u keybase bash << EOF
source ./env.sh
nohup bash -c "KEYBASE_RUN_MODE=prod kbfsfuse /keybase | grep -v 'ERROR Mounting the filesystem failed' &"
sleep 3
keybase oneshot --username \$KEYBASE_USERNAME --paperkey "\$KEYBASE_PAPERKEY"
bin/keybaseca service
EOF