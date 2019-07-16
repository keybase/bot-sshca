#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# Some colors for pretty output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# A function used to indent the log output from the tests
indent() { sed 's/^/    /'; }

cd tests/simple/
../reset.sh
source env.sh
cat keybaseca.config.gen | envsubst > keybaseca.config
echo "Building containers..."
docker-compose build 2>&1 > /dev/null
echo "Running integration tests..."
docker-compose up -d

TEST_EXIT_CODE=`docker wait kssh`

docker logs kssh | indent

if [ -z ${TEST_EXIT_CODE+x} ] || [ "$TEST_EXIT_CODE" -ne 0 ] ; then
  printf "${RED}Simple Mode Tests Failed${NC} - Exit Code: $TEST_EXIT_CODE\n"
else
  if (docker logs kssh | python3 ../integrationTestUtils.py count 8) ; then
    printf "${GREEN}Simple Mode Tests Passed${NC}\n"
  else
    printf "${RED}Simple Mode Tests Missing Output${NC}\n"
  fi
fi

docker-compose stop 2>&1 > /dev/null
docker-compose kill 2>&1 > /dev/null
docker-compose rm -f

cd ../advanced/
../reset.sh
source env.sh
cat keybaseca.config.gen | envsubst > keybaseca.config
echo "Building containers..."
docker-compose build 2>&1 > /dev/null
echo "Running integration tests..."
docker-compose up -d

TEST_EXIT_CODE=`docker wait kssh`

docker logs kssh | indent

if [ -z ${TEST_EXIT_CODE+x} ] || [ "$TEST_EXIT_CODE" -ne 0 ] ; then
  printf "${RED}Advanced Mode Tests Failed${NC} - Exit Code: $TEST_EXIT_CODE\n"
else
  if (docker logs kssh | python3 ../integrationTestUtils.py count 11) ; then
    printf "${GREEN}Advanced Mode Tests Passed${NC}\n"
  else
    printf "${RED}Advanced Mode Tests Missing Output${NC}\n"
  fi
fi

docker-compose stop 2>&1 > /dev/null
docker-compose kill 2>&1 > /dev/null
docker-compose rm -f

../reset.sh
