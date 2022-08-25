#!/usr/bin/env bash

LAST_WD=$(pwd)

echo "Start mpc-server"

read LAST_TEST_WD < $HOME/mpctest/testwd_last

cd $LAST_TEST_WD/mpc-server

mkdir secret
mkdir logs
mkdir db

MPC_SERVER_REPO=$HOME/mpctest/mpc-server

echo -n "f6826fc16547130848ea32196c95457b4698feded1a8f109eb224ddcd27d66af7d0f88b44d2765a4a567a8999a4410852510deacdcb87e0c7cfe23fa1d0090a8" > secret/p1.s
echo -n "9f099ed7a7615ce2d7de100a8feaf39adfa904146ee523f6dacaad6a3f69b4e9d10f2b01c13dc31b1265f9042d7437f63ad81c5113dec541bb799e49dd21c571" > secret/p2.s
echo -n "e3858ec05c7762d79de30428a2561a520123b1e2b16687ae57ba0ba550e07ffac49861bab15fe62678435c8ae2a57522752a7b8af68d7591ca0e7b44c5f64a4d" > secret/p3.s
echo -n "a8f68535da081ab3818d9f00f64730583b4d31b918b44771526c925d833df96b8c258fff5e730f6c806a99edc24ad90ef177c0918517527f81e908ec7009feaa" > secret/p4.s
echo -n "7aa717fdbb98618c438224df53b1de435932a831f430495a06f00b5c0aabfc63eaf7e9439361a46ecac4a42b5ba39779373c57d6c45ac9ac06b7bd42e03abaf8" > secret/p5.s
echo -n "5542f197e10c3b91c05767fe0906d7caad6380913df0f9b850f0044dd61b3b7358dd8db2e5f4961c2f799d1442f095b36ce12b1011861948e72575877ead7f7b" > secret/p6.s
echo -n "0c56c3f0d5ae14d8853e964430f517b8016ac7e24dd15179d5d2657c38e7d6fa285ab0fcc748bd8c822a65e02f5a07d7e3be7ef3980bfd382bbda1bf5477fc3d" > secret/p7.s

pwd="RBuCJbmWY1Mtcl5LoMRqkQQpT5GJmCEvbuRR7ewCPDATBzFtm9a6jhIovftgddmL"

RUST_BACKTRACE=full $MPC_SERVER_REPO/messenger/target/debug/messenger > logs/messenger.log 2>&1 &
RUST_BACKTRACE=full $MPC_SERVER_REPO/target/debug/mpc-server -s secret/p1.s --password $pwd --port 8001 -m http://127.0.0.1:8000  --db-path db/p1.db > logs/p1.log 2>&1 &
RUST_BACKTRACE=full $MPC_SERVER_REPO/target/debug/mpc-server -s secret/p2.s --password $pwd --port 8002 -m http://127.0.0.1:8000  --db-path db/p2.db > logs/p2.log 2>&1 &
RUST_BACKTRACE=full $MPC_SERVER_REPO/target/debug/mpc-server -s secret/p3.s --password $pwd --port 8003 -m http://127.0.0.1:8000  --db-path db/p3.db > logs/p3.log 2>&1 &
RUST_BACKTRACE=full $MPC_SERVER_REPO/target/debug/mpc-server -s secret/p4.s --password $pwd --port 8004 -m http://127.0.0.1:8000  --db-path db/p4.db > logs/p4.log 2>&1 &
RUST_BACKTRACE=full $MPC_SERVER_REPO/target/debug/mpc-server -s secret/p5.s --password $pwd --port 8005 -m http://127.0.0.1:8000  --db-path db/p5.db > logs/p5.log 2>&1 &
RUST_BACKTRACE=full $MPC_SERVER_REPO/target/debug/mpc-server -s secret/p6.s --password $pwd --port 8006 -m http://127.0.0.1:8000  --db-path db/p6.db > logs/p6.log 2>&1 &
RUST_BACKTRACE=full $MPC_SERVER_REPO/target/debug/mpc-server -s secret/p7.s --password $pwd --port 8007 -m http://127.0.0.1:8000  --db-path db/p7.db > logs/p7.log 2>&1 &

cd $LAST_WD