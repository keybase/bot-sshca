#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

docker system prune -f
docker volume ls -q | xargs -r -- docker volume rm -f
