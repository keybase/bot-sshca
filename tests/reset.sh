#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

docker system prune -f
docker-compose down -v
