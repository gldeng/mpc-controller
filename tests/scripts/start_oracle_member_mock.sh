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

$oracle_dir/oracle_service_mock --oracleMemberPK a54a5d692d239287e8358f27caee92ab5756c0276a6db0a062709cd86451a855 -oracleManagerAddr $ORACLE_MANAGER_ADDRESS > logs/oracle1.log 2>&1 &
$oracle_dir/oracle_service_mock --oracleMemberPK 86a5e025e16a96e2706d72fd6115f2ee9ae1c5dfc4c53894b70b19e6fc73b838 -oracleManagerAddr $ORACLE_MANAGER_ADDRESS > logs/oracle2.log 2>&1 &
$oracle_dir/oracle_service_mock --oracleMemberPK d876abc4ef78972fc733651bfc79676d9a6722626f9980e2db249c22ed57dbb2 -oracleManagerAddr $ORACLE_MANAGER_ADDRESS > logs/oracle3.log 2>&1 &
$oracle_dir/oracle_service_mock --oracleMemberPK 6353637e9d5cdc0cbc921dadfcc8877d54c0a05b434a1d568423cb918d582eac -oracleManagerAddr $ORACLE_MANAGER_ADDRESS > logs/oracle4.log 2>&1 &
$oracle_dir/oracle_service_mock --oracleMemberPK c847f461acdd47f2f0bf08b7480d68f940c97bbc6c0a5a03e0cbefae4d9a7592 -oracleManagerAddr $ORACLE_MANAGER_ADDRESS > logs/oracle5.log 2>&1 &

cd $LAST_WD
