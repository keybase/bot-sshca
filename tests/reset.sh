#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

docker system prune -f

OUTPUT=$(docker volume ls -q)
if [[ $OUTPUT ]]; then
    echo $OUTPUT | xargs -r -- docker volume rm -f
fi
