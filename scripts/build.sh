#!/usr/bin/env bash

# Build avalanchego

cd ../avalanchego

if [ ! -d "build" ]; then
  chmod +x ./scripts/build.sh
  echo "Start building avalanchego..."
  ./scripts/build.sh
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

# Build mpc-controller
cd ../mpc-controller
if [ ! -f "mpc-controller" ]; then
  echo "Start building mpc-controller..."
  go build
fi
