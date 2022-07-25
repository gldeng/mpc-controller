#!/usr/bin/env bash

LAST_WD=$(pwd)

mkdir -p $HOME/mpctest/mpc-server-mock

# Build mpc-controller
if [ ! -f "$HOME/mpctest/mpc-server-mock/mpc-server-mock" ]; then
  echo "Start building mpc-server-mock..."
  cd ./tests/mocks/mpc_server/
  go clean
  go build -o mpc-server-mock
  mv mpc-server-mock $HOME/mpctest/mpc-server-mock/
  cd $LAST_WD
fi

echo "Start mpc-server-mock"
read LAST_TEST_WD < $HOME/mpctest/testwd_last
mkdir -p $LAST_TEST_WD/mpc-server-mock
cd $LAST_TEST_WD/mpc-server-mock

MPC_SERVER_MOCK_REPO=$HOME/mpctest/mpc-server-mock

$MPC_SERVER_MOCK_REPO/mpc-server-mock -p 7 -t 4 > log.log 2>&1 &

cd $LAST_WD