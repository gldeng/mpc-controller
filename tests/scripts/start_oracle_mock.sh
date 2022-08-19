#!/usr/bin/env bash

LAST_WD=$(pwd)

mkdir -p $HOME/mpctest/oracle_service

# Build Oracle service mock
if [ ! -f "$HOME/mpctest/oracle_service/oracle_service_mock" ]; then
  echo "Start building oracle_service_mock..."
  cd ./tests/mocks/oracle_service/
  go clean
  go build -o oracle_service_mock
  mv oracle_service_mock $HOME/mpctest/oracle_service/
  cd $LAST_WD
fi

echo "Start oracle_service_mock"
read LAST_TEST_WD < $HOME/mpctest/testwd_last
mkdir -p $LAST_TEST_WD/oracle_service_mock
cd $LAST_TEST_WD/oracle_service_mock

ORACLE_MANAGER_ADDRESS=$(cat $HOME/mpctest/contracts/addresses/ORACLE_MANAGER_ADDRESS)

oracle_dir=$HOME/mpctest/oracle_service

$oracle_dir/oracle_service_mock -oracleManagerAddr $ORACLE_MANAGER_ADDRESS > log.log 2>&1 &

cd $LAST_WD