#!/usr/bin/env bash

PROCESS_AVALANCHE=$(ps -aux | grep avalanchego | xargs | wc -l)
if [ $PROCESS_AVALANCHE -gt 1 ]; then
  echo "Killing avalanchego"
  pkill -f avalanchego
fi

PROCESS_MPC_CONTROLLER=$(ps -aux | grep mpc-controller | xargs | wc -l)
if [ $PROCESS_MPC_CONTROLLER -gt 1 ]; then
  echo "Killing mpc-controller"
  pkill -f mpc-controller
fi

PROCESS_MPC_SERVER=$(ps -aux | grep mpc-server | xargs | wc -l)
if [ $PROCESS_MPC_SERVER -gt 1 ]; then
  echo "Killing mpc-server"
  pkill -f mpc-server
fi


