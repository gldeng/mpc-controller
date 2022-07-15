#!/usr/bin/env bash

LAST_WD=$(pwd)

echo "Start avalanchego"

read LAST_TEST_WD < $HOME/mpctest/testwd_last

cd $LAST_TEST_WD/avalanchego

mkdir log
mkdir db

AVALANCHEGO_REPO=$HOME/mpctest/avalanchego

$AVALANCHEGO_REPO/build/avalanchego --log-level=debug --public-ip=127.0.0.1 --http-port=9650 --staking-port=9651 --db-dir=db/node1 --network-id=local --staking-tls-cert-file=$AVALANCHEGO_REPO/staking/local/staker1.crt --staking-tls-key-file=$AVALANCHEGO_REPO/staking/local/staker1.key > log/node1.log 2>&1 &

$AVALANCHEGO_REPO/build/avalanchego --log-level=debug --public-ip=127.0.0.1 --http-port=9652 --staking-port=9653 --db-dir=db/node2 --network-id=local --bootstrap-ips=127.0.0.1:9651 --bootstrap-ids=NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg --staking-tls-cert-file=$AVALANCHEGO_REPO/staking/local/staker2.crt --staking-tls-key-file=$AVALANCHEGO_REPO/staking/local/staker2.key > log/node2.log 2>&1 &

$AVALANCHEGO_REPO/build/avalanchego --log-level=debug --public-ip=127.0.0.1 --http-port=9654 --staking-port=9655 --db-dir=db/node3 --network-id=local --bootstrap-ips=127.0.0.1:9651 --bootstrap-ids=NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg --staking-tls-cert-file=$AVALANCHEGO_REPO/staking/local/staker3.crt --staking-tls-key-file=$AVALANCHEGO_REPO/staking/local/staker3.key > log/node3.log 2>&1 &

cd $LAST_WD
