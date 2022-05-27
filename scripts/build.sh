#!/usr/bin/env bash

avalanche="git@github.com:ava-labs/avalanchego.git"
mpcServer="git@github.com:AvaLido/mpc-server.git"

if [ ! -d "avalanchego" ]; then
  git clone $avalanche
  cd avalanchego
  ./scripts/build.sh
  cd ..
fi

if [ ! -d "mpc-server" ]; then
  git clone $mpcServer
  cd mpc-server
  cd messenger
  cargo build
  cd ../secp256k1-id
  cargo build
  cd ..
  cargo build
  cd ..
fi

cd ..
go build

cd ./tests

sleep 10
