#!/usr/bin/env bash

avalanche=$(ps aux | grep avalanchego)
if [ ! -d "$avalanche" ]; then
  pkill -f avalanchego
fi

mpcController=$(ps aux | grep mpc-controller)
if [ ! -d "$mpcController" ]; then
  pkill -f mpc-controller
fi

mpcServer=$(ps aux | grep mpc-server)
if [ ! -d "$mpcServer" ]; then
  pkill -f mpc-server
fi


