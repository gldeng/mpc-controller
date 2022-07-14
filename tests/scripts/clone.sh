#!/usr/bin/env bash

mkdir -p /tmp/mpctest

mkdir -p /tmp/mpctest/mpc-controller

if [ ! -d "/tmp/mpctest/avalanchego" ]; then
  git clone git@github.com:ava-labs/avalanchego.git /tmp/mpctest/avalanchego
  cp ./tests/configs/genesis/genesis_local.go /tmp/mpctest/avalanchego/genesis/genesis_local.go
fi

if [ ! -d "/tmp/mpctest/mpc-server" ]; then
  git clone git@github.com:AvaLido/mpc-server.git /tmp/mpctest/mpc-server
fi

if [ ! -d "/tmp/mpctest/contracts" ]; then
  git submodule init
  git submodule update

  LAST_WD=$(pwd)
  cd /tmp/mpctest/

  forge init contracts
  cp -a $LAST_WD/contract/src/. contracts/src/
  cp -a $LAST_WD/contract/lib/. contracts/lib/

  rm contracts/src/Contract.sol contracts/test/Contract.t.sol
  cd $LAST_WD
fi