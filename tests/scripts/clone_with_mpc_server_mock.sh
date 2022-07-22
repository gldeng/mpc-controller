#!/usr/bin/env bash

mkdir -p $HOME/mpctest
mkdir -p $HOME/mpctest/mpc-controller
mkdir -p $HOME/mpctest/mpc-server-mock

if [ ! -d "$HOME/mpctest/avalanchego" ]; then
  git clone git@github.com:ava-labs/avalanchego.git $HOME/mpctest/avalanchego
  cp ./tests/configs/genesis/genesis_local.go $HOME/mpctest/avalanchego/genesis/genesis_local.go
fi

#if [ ! -d "$HOME/mpctest/mpc-server" ]; then
#  git clone git@github.com:AvaLido/mpc-server.git $HOME/mpctest/mpc-server
#fi

if [ ! -d "$HOME/mpctest/contracts" ]; then
  git submodule init
  git submodule update

  LAST_WD=$(pwd)
  cd $HOME/mpctest/

  forge init contracts
  cp -a $LAST_WD/contract/src/. contracts/src/
  cp -a $LAST_WD/contract/lib/. contracts/lib/

  rm contracts/src/Contract.sol contracts/test/Contract.t.sol
  cd $LAST_WD
fi