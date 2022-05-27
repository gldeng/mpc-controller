#!/usr/bin/env bash

address=$1
ip=localhost

./mpc-controller --configFile ./config/config1.yaml  2>&1 | tee p1.log &
./mpc-controller --configFile ./config/config2.yaml  2>&1 | tee p2.log &
./mpc-controller --configFile ./config/config3.yaml  2>&1 | tee p3.log &
