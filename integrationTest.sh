#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# Some colors for pretty output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# A function used to indent the log output from the tests
indent() { sed 's/^/    /'; }

cd tests/
source env.sh
cat keybaseca.config.gen | envsubst > keybaseca.config
echo "Building containers..."
docker-compose build 2>&1 > /dev/null
echo "Running integration tests..."
docker-compose up -d

TEST_EXIT_CODE=`docker wait docker_kssh_1`

docker logs docker_kssh_1 | indent

if [ -z ${TEST_EXIT_CODE+x} ] || [ "$TEST_EXIT_CODE" -ne 0 ] ; then
  printf "${RED}Tests Failed${NC} - Exit Code: $TEST_EXIT_CODE\n"
else
  printf "${GREEN}Tests Passed${NC}\n"
fi

docker-compose stop 2>&1 > /dev/null
docker-compose kill 2>&1 > /dev/null
docker-compose rm -f