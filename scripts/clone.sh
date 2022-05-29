#!/usr/bin/env bash

cd ..

if [ ! -d "avalanchego" ]; then
  git clone git@github.com:ava-labs/avalanchego.git
fi

if [ ! -d "mpc-server" ]; then
  git clone git@github.com:AvaLido/mpc-server.git
fi