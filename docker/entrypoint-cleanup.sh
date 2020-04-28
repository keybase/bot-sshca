#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# chown as root
chown -R keybase:keybase /mnt

# Run everything else as the keybase user
sudo -i -u keybase bash << EOF
export "FORCE_WRITE=$FORCE_WRITE"
export "KEYBASE_USERNAME=$KEYBASE_USERNAME"
export "KEYBASE_PAPERKEY=$KEYBASE_PAPERKEY"
nohup bash -c "KEYBASE_RUN_MODE=prod kbfsfuse /keybase | grep -v 'ERROR Mounting the filesystem failed' &"
sleep ${KEYBASE_TIMEOUT:-5}
keybase oneshot
bin/keybaseca --wipe-all-configs
sleep ${KEYBASE_TIMEOUT:-5}
EOF
