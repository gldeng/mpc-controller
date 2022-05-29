#!/usr/bin/env bash

PROCESS_AVALANCHE=$(ps -aux | grep avalanchego | wc -l)
if [ $PROCESS_AVALANCHE -gt 1 ]; then
  echo "Killing avalanchego"
  pkill -f avalanchego
fi

# todo: deal with: task: Failed to run task "tests:kill": exit status 143
PROCESS_MPC_CONTROLLER=$(ps -aux | grep mpc-controller | wc -l)
if [ $PROCESS_MPC_CONTROLLER -gt 3 ]; then
  echo "Killing mpc-controller"
  pkill -f mpc-controller
fi

PROCESS_MPC_SERVER=$(ps -aux | grep mpc-server | wc -l)
if [ $PROCESS_MPC_SERVER -gt 1 ]; then
  echo "Killing mpc-server"
  pkill -f mpc-server
fi

PROCESS_MPC_MESSENGER=$(ps -aux | grep messenger | wc -l)
if [ $PROCESS_MPC_MESSENGER -gt 3 ]; then
  echo "Killing mpc-messenger"
  pkill -f messenger
fi

sleep 3