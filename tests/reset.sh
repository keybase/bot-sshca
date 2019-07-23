#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

docker-compose down -v 2>&1 > /dev/null
docker system prune -f 2>&1 > /dev/null

OUTPUT=$(docker volume ls -q)
if [[ $OUTPUT ]]; then
    echo $OUTPUT | xargs docker volume rm -f 2>&1 > /dev/null
fi
