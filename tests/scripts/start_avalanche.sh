#!/usr/bin/env bash

LAST_WD=$(pwd)

echo "Starting avalanchego"

read LAST_TEST_WD < /tmp/mpctest/testwd_last

cd $LAST_TEST_WD/avalanchego

mkdir log
mkdir db

AVALANCHEGO_REPO=/tmp/mpctest/avalanchego/

$AVALANCHEGO_REPO/build/avalanchego --public-ip=127.0.0.1 --http-port=9650 --staking-port=9651 --db-dir=db/node1 --network-id=local --staking-tls-cert-file=$AVALANCHEGO_REPO/staking/local/staker1.crt --staking-tls-key-file=$AVALANCHEGO_REPO/staking/local/staker1.key  2>&1 | tee testdir/log/node1.log &

$AVALANCHEGO_REPO/build/avalanchego --public-ip=127.0.0.1 --http-port=9652 --staking-port=9653 --db-dir=db/node2 --network-id=local --bootstrap-ips=127.0.0.1:9651 --bootstrap-ids=NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg --staking-tls-cert-file=$AVALANCHEGO_REPO/staking/local/staker2.crt --staking-tls-key-file=$AVALANCHEGO_REPO/staking/local/staker2.key  2>&1 | tee log/node2.log &

$AVALANCHEGO_REPO/build/avalanchego --public-ip=127.0.0.1 --http-port=9654 --staking-port=9655 --db-dir=db/node3 --network-id=local --bootstrap-ips=127.0.0.1:9651 --bootstrap-ids=NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg --staking-tls-cert-file=$AVALANCHEGO_REPO/staking/local/staker3.crt --staking-tls-key-file=$AVALANCHEGO_REPO/staking/local/staker3.key  2>&1 | tee log/node3.log &

$AVALANCHEGO_REPO/build/avalanchego --public-ip=127.0.0.1 --http-port=9656 --staking-port=9657 --db-dir=db/node4 --network-id=local --bootstrap-ips=127.0.0.1:9651 --bootstrap-ids=NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg --staking-tls-cert-file=$AVALANCHEGO_REPO/staking/local/staker4.crt --staking-tls-key-file=$AVALANCHEGO_REPO/staking/local/staker4.key  2>&1 | tee log/node4.log &

$AVALANCHEGO_REPO/build/avalanchego --public-ip=127.0.0.1 --http-port=9658 --staking-port=9659 --db-dir=db/node5 --network-id=local --bootstrap-ips=127.0.0.1:9651 --bootstrap-ids=NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg --staking-tls-cert-file=$AVALANCHEGO_REPO/staking/local/staker5.crt --staking-tls-key-file=$AVALANCHEGO_REPO/staking/local/staker5.key  2>&1 | tee log/node5.log &

cd $LAST_WD
