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

mkdir -p logs

$oracle_dir/oracle_service_mock --oracleMemberPK 0x03C1196617387899390d3a98fdBdfD407121BB67 -oracleManagerAddr $ORACLE_MANAGER_ADDRESS > logs/oracle1.log 2>&1 &
$oracle_dir/oracle_service_mock --oracleMemberPK 0x6C58f6E7DB68D9F75F2E417aCbB67e7Dd4e413bf -oracleManagerAddr $ORACLE_MANAGER_ADDRESS > logs/oracle2.log 2>&1 &
$oracle_dir/oracle_service_mock --oracleMemberPK 0xa7bB9405eAF98f36e2683Ba7F36828e260BD0018 -oracleManagerAddr $ORACLE_MANAGER_ADDRESS > logs/oracle3.log 2>&1 &
$oracle_dir/oracle_service_mock --oracleMemberPK 0xE339767906891bEE026285803DA8d8F2f346842C -oracleManagerAddr $ORACLE_MANAGER_ADDRESS > logs/oracle4.log 2>&1 &
$oracle_dir/oracle_service_mock --oracleMemberPK 0x0309a747a34befD1625b5dcae0B00625FAa30460 -oracleManagerAddr $ORACLE_MANAGER_ADDRESS > logs/oracle5.log 2>&1 &

cd $LAST_WD
