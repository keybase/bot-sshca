#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# Unit tests first
go test ./... 2>&1 | grep -v 'no test files'

if [[ -f "tests/env.sh" ]]; then
    echo "env.sh file already exists, skipping configuring new accounts..."
else
    python3 tests/configure_accounts.py
fi


# Some colors for pretty output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# A function used to indent the log output from the tests
indent() { sed 's/^/    /'; }
reset_docker() {
    docker-compose down -v
    docker system prune -f
}

cd tests/
source env.sh
reset_docker

echo "Building containers..."
cd ../docker/ && make && cd ../tests/
docker-compose build
echo "Running integration tests..."
docker-compose up -d

docker logs kssh -f | indent
TEST_EXIT_CODE=`docker wait kssh`

if [ -z ${TEST_EXIT_CODE+x} ] || [ "$TEST_EXIT_CODE" -ne 0 ] ; then
  printf "${RED}Tests Failed${NC} - Exit Code: $TEST_EXIT_CODE\n"
else
  printf "${GREEN}Tests Passed${NC}\n"
fi

docker-compose stop 2>&1 > /dev/null
docker-compose kill 2>&1 > /dev/null
docker-compose rm -f

reset_docker
