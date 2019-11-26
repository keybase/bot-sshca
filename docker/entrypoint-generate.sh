#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# chown as root
chown -R keybase:keybase /mnt

# Run everything else as the keybase user
sudo -i -u keybase bash << EOF
source ./env.sh
export "FORCE_WRITE=$FORCE_WRITE"
keybase oneshot --username \$KEYBASE_USERNAME --paperkey "\$KEYBASE_PAPERKEY"
bin/keybaseca generate
EOF