#!/usr/bin/env bash

LAST_WD=$(pwd)

echo "Start mpc-server"

read LAST_TEST_WD < /tmp/mpctest/testwd_last

cd $LAST_TEST_WD/mpc-server

mkdir secret
mkdir log
mkdir db

MPC_SERVER_REPO=/tmp/mpctest/mpc-server/

echo -n "59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21" > secret/p1.s
echo -n "6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33" > secret/p2.s
echo -n "5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b" > secret/p3.s

RUST_BACKTRACE=full $MPC_SERVER_REPO/messenger/target/debug/messenger > log/messenger.log 2>&1 &
RUST_BACKTRACE=full $MPC_SERVER_REPO/target/debug/mpc-server -s secret/p1.s --port 8001 -m http://127.0.0.1:8000  --db-path db/p1.db > log/p1.log 2>&1 &
RUST_BACKTRACE=full $MPC_SERVER_REPO/target/debug/mpc-server -s secret/p2.s --port 8002 -m http://127.0.0.1:8000  --db-path db/p2.db > log/p2.log 2>&1 &
RUST_BACKTRACE=full $MPC_SERVER_REPO/target/debug/mpc-server -s secret/p3.s --port 8003 -m http://127.0.0.1:8000  --db-path db/p3.db > log/p3.log 2>&1 &

cd $LAST_WD