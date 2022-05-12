address=$1
ip=localhost

pkill -f mpc-controller

sleep 5

./mpc-controller --configFile ./config1.yaml  2>&1 | tee p1.log &
echo "Mpc-controller 1 started"

./mpc-controller --configFile ./config2.yaml  2>&1 | tee p1.log &
echo "Mpc-controller 2 started"

./mpc-controller --configFile ./config3.yaml  2>&1 | tee p1.log &
echo "Mpc-controller 3 started"
