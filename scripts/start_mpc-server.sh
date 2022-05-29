#!/usr/bin/env bash

PROCESS_MPC_SERVER=$(ps -aux | grep mpc-server | wc -l)
if [ ! $PROCESS_MPC_SERVER -gt 3 ]; then
  echo "Starting mpc-server"

  cd ../mpc-server

  mkdir mpc
  cd mpc
  mkdir secret
  mkdir log
  mkdir db
  cd ..

  echo -n "59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21" > mpc/secret/p1.s
  echo -n "6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33" > mpc/secret/p2.s
  echo -n "5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b" > mpc/secret/p3.s

  RUST_BACKTRACE=full ./messenger/target/debug/messenger 2>&1 | tee mpc/log/messenger.log &
  RUST_BACKTRACE=full ./target/debug/mpc-server -s mpc/secret/p1.s --port 8001 -m http://127.0.0.1:8000  --db-path mpc/db/p1.db 2>&1 | tee mpc/log/p1.log &
  RUST_BACKTRACE=full ./target/debug/mpc-server -s mpc/secret/p2.s --port 8002 -m http://127.0.0.1:8000  --db-path mpc/db/p2.db 2>&1 | tee mpc/log/p2.log &
  RUST_BACKTRACE=full ./target/debug/mpc-server -s mpc/secret/p3.s --port 8003 -m http://127.0.0.1:8000  --db-path mpc/db/p3.db 2>&1 | tee mpc/log/p3.log &
  sleep 3
  cd ../mpc-controller
fi
