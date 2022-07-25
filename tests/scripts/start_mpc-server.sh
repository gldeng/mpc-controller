#!/usr/bin/env bash

LAST_WD=$(pwd)

echo "Start mpc-server"

read LAST_TEST_WD < $HOME/mpctest/testwd_last

cd $LAST_TEST_WD/mpc-server

mkdir secret
mkdir logs
mkdir db

MPC_SERVER_REPO=$HOME/mpctest/mpc-server

echo -n "59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21" > secret/p1.s
echo -n "6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33" > secret/p2.s
echo -n "5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b" > secret/p3.s
echo -n "156177364ae1ca503767382c1b910463af75371856e90202cb0d706cdce53c33" > secret/p4.s
echo -n "353fb105bbf9c29cbf46d4c93a69587ac478138b7715f0786d7ae1cc05230878" > secret/p5.s
echo -n "b17eac91d7aa2bd5fa72916b6c8a35ab06e8f0c325c98067bbc9645b85ce789f" > secret/p6.s
echo -n "7084300e7059ea4b308ec5b965ef581d3f9c9cd63714082ccf9b9d1fb34d658b" > secret/p7.s

RUST_BACKTRACE=full $MPC_SERVER_REPO/messenger/target/debug/messenger > log/messenger.log 2>&1 &
RUST_BACKTRACE=full $MPC_SERVER_REPO/target/debug/mpc-server -s secret/p1.s --port 8001 -m http://127.0.0.1:8000  --db-path db/p1.db > logs/p1.log 2>&1 &
RUST_BACKTRACE=full $MPC_SERVER_REPO/target/debug/mpc-server -s secret/p2.s --port 8002 -m http://127.0.0.1:8000  --db-path db/p2.db > logs/p2.log 2>&1 &
RUST_BACKTRACE=full $MPC_SERVER_REPO/target/debug/mpc-server -s secret/p3.s --port 8003 -m http://127.0.0.1:8000  --db-path db/p3.db > logs/p3.log 2>&1 &
RUST_BACKTRACE=full $MPC_SERVER_REPO/target/debug/mpc-server -s secret/p4.s --port 8004 -m http://127.0.0.1:8000  --db-path db/p4.db > logs/p4.log 2>&1 &
RUST_BACKTRACE=full $MPC_SERVER_REPO/target/debug/mpc-server -s secret/p5.s --port 8005 -m http://127.0.0.1:8000  --db-path db/p5.db > logs/p5.log 2>&1 &
RUST_BACKTRACE=full $MPC_SERVER_REPO/target/debug/mpc-server -s secret/p6.s --port 8006 -m http://127.0.0.1:8000  --db-path db/p6.db > logs/p6.log 2>&1 &
RUST_BACKTRACE=full $MPC_SERVER_REPO/target/debug/mpc-server -s secret/p7.s --port 8007 -m http://127.0.0.1:8000  --db-path db/p7.db > logs/p7.log 2>&1 &

cd $LAST_WD