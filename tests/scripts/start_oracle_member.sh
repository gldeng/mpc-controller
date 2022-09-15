#!/usr/bin/env bash

# Network URLs
C_CHAIN_RPC_URL="http://34.172.25.188:9650"

LAST_WD=$(pwd)

# Clone Oracle repository
if [ ! -d "$HOME/mpctest/oracle" ]; then
  mkdir -p $HOME/mpctest/oracle
  LAST_WD=$(pwd)
  cd $HOME/mpctest/
  git clone git@github.com:AvaLido/oracle.git
  cd oracle
  git submodule update --init --recursive --remote
  cd $LAST_WD
fi

# Build Oracle service
if [ ! -f "$HOME/mpctest/oracle/oracle" ]; then
  echo "Start building oracle..."
  LAST_WD=$(pwd)
  cd $HOME/mpctest/oracle
  go clean
  go build -o oracle
  cd $LAST_WD
fi

echo "Start Oracle service"
read LAST_TEST_WD < $HOME/mpctest/testwd_last
mkdir -p $LAST_TEST_WD/oracle
cd $LAST_TEST_WD/oracle

ORACLE_ADDRESS=$(cat $HOME/mpctest/contracts/addresses/ORACLE_ADDRESS)
ORACLE_MANAGER_ADDRESS=$(cat $HOME/mpctest/contracts/addresses/ORACLE_MANAGER_ADDRESS)

oracle_dir=$HOME/mpctest/oracle

mkdir -p logs

$oracle_dir/oracle --rpc-url $C_CHAIN_RPC_URL --max-delegation-fee 10000 --min-uptime 1209600 --max-stake-period 9500 --oracle-address $ORACLE_ADDRESS --oracle-manager-address $ORACLE_MANAGER_ADDRESS --private-key a54a5d692d239287e8358f27caee92ab5756c0276a6db0a062709cd86451a855 > logs/oracle1.log 2>&1 &
$oracle_dir/oracle --rpc-url $C_CHAIN_RPC_URL --max-delegation-fee 10000 --min-uptime 1209600 --max-stake-period 9500 --oracle-address $ORACLE_ADDRESS --oracle-manager-address $ORACLE_MANAGER_ADDRESS --private-key 86a5e025e16a96e2706d72fd6115f2ee9ae1c5dfc4c53894b70b19e6fc73b838 > logs/oracle2.log 2>&1 &
$oracle_dir/oracle --rpc-url $C_CHAIN_RPC_URL --max-delegation-fee 10000 --min-uptime 1209600 --max-stake-period 9500 --oracle-address $ORACLE_ADDRESS --oracle-manager-address $ORACLE_MANAGER_ADDRESS --private-key d876abc4ef78972fc733651bfc79676d9a6722626f9980e2db249c22ed57dbb2 > logs/oracle3.log 2>&1 &
$oracle_dir/oracle --rpc-url $C_CHAIN_RPC_URL --max-delegation-fee 10000 --min-uptime 1209600 --max-stake-period 9500 --oracle-address $ORACLE_ADDRESS --oracle-manager-address $ORACLE_MANAGER_ADDRESS --private-key 6353637e9d5cdc0cbc921dadfcc8877d54c0a05b434a1d568423cb918d582eac > logs/oracle4.log 2>&1 &
$oracle_dir/oracle --rpc-url $C_CHAIN_RPC_URL --max-delegation-fee 10000 --min-uptime 1209600 --max-stake-period 9500 --oracle-address $ORACLE_ADDRESS --oracle-manager-address $ORACLE_MANAGER_ADDRESS --private-key c847f461acdd47f2f0bf08b7480d68f940c97bbc6c0a5a03e0cbefae4d9a7592 > logs/oracle5.log 2>&1 &

cd $LAST_WD
