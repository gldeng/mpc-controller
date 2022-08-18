#!/usr/bin/env bash

mkdir -p $HOME/mpctest
mkdir -p $HOME/mpctest/mpc-controller

if [ ! -d "$HOME/mpctest/avalanchego" ]; then
  git clone git@github.com:ava-labs/avalanchego.git $HOME/mpctest/avalanchego
  LAST_WD=$(pwd)
  cd $HOME/mpctest/avalanchego
  git checkout tags/v1.7.14
  cd $LAST_WD
  cp ./tests/configs/genesis/genesis_local.go $HOME/mpctest/avalanchego/genesis/genesis_local.go
fi

if [ ! -d "$HOME/mpctest/mpc-server" ]; then
  git clone git@github.com:AvaLido/mpc-server.git $HOME/mpctest/mpc-server
fi

if [ ! -d "$HOME/mpctest/contracts" ]; then
  LAST_WD=$(pwd)
  cd $HOME/mpctest/
  git clone git@github.com:AvaLido/contracts.git
  cd contracts
  git submodule update --init --recursive --remote
  cp -n $LAST_WD/tests/contracts/deploy/Deploy.t.sol src/deploy/Deploy.t.sol
  cd $LAST_WD
fi