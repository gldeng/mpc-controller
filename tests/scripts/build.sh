#!/usr/bin/env bash

# Build mpc-controller
if [ ! -f "$HOME/mpctest/mpc-controller/mpc-controller" ]; then
  echo "Start building mpc-controller..."
  cd ./cmd/mpc-controller
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

## Note: install libgmp-dev and libssl-dev before building mpc-server
#sudo apt-get install libgmp-dev
#sudo apt-get install libssl-dev

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