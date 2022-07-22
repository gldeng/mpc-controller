#!/usr/bin/env bash

# Build mpc-controller
if [ ! -f "$HOME/mpctest/mpc-controller/mpc-controller" ]; then
  echo "Start building mpc-controller..."
  cd ./cmd/mpc-controller
  go clean
  go build
  mv mpc-controller $HOME/mpctest/mpc-controller/
  cd ../../
fi

LAST_WD=$(pwd)

cd $HOME/mpctest/

# Build avalanchego

cd avalanchego/

if [ ! -d "build" ]; then
  echo "Start building avalanchego..."
  bash ./scripts/build.sh
fi

# Build mpc-server

cd ../mpc-server
cd messenger
if [ ! -d "target" ]; then
  echo "Start building mpc-server/messenger..."
  cargo build
fi

cd ../secp256k1-id
if [ ! -d "target" ]; then
  echo "Start building mpc-server/secp256k1-id..."
  cargo build
fi

cd ..
if [ ! -d "target" ]; then
  echo "Start building mpc-server..."
  cargo build
fi

cd $LAST_WD

# Build mpc-server mock

if [ ! -f "$HOME/mpctest/mpc-server-mock/mpc_server" ]; then
  echo "Start building mpc-server-mock..."
  cd ./tests/mocks/mpc_server/
  go clean
  go build
  mv mpc_server $HOME/mpctest/mpc-server-mock/
  cd $LAST_WD
fi
