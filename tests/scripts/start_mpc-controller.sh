#!/usr/bin/env bash

rm -rf dbs
rm -rf configs
rm -rf logs

# todo: use mktemp instead
mkdir -p dbs
mkdir -p configs
mkdir -p logs

sks=("59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21" "6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33" "5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b")
MPC_SERVER_URLS=("http://localhost:8001" "http://localhost:8002" "http://localhost:8003")

MPC_MANAGER_ADDRESS=$(cat addresses/MPC_MANAGER_ADDRESS)
function create_config(){
    id=$1
    sk=${sks[$(expr ${id} - 1)]}
    mpcServerUrl=${MPC_SERVER_URLS[$(expr ${id} - 1)]}
    read -r -d '' CFG <<- EOM
enableDevMode: true
controllerId: "mpc-controller-0${id}"
controllerKey: "${sk}"
coordinatorAddress: "${MPC_MANAGER_ADDRESS}"
mpcServerUrl: "${mpcServerUrl}"
ethRpcUrl: "http://localhost:9650/ext/bc/C/rpc"
ethWsUrl: "ws://127.0.0.1:9650/ext/bc/C/ws"
cChainIssueUrl: "http://localhost:9650"
pChainIssueUrl: "http://localhost:9650"
confignetwork:
  networkId: 12345
  chainId: 43112
  cChainId: "2CA6j5zYzasynPsFeNoqWkmTCt3VScMvXUZHbfDJ8k3oGzAPtU"
  avaxId: "2fombhL7aGPwj3KH4bfrmJwW6PVnMobf9Y2fn9GwxiAAJyFDbe"
  importFee: 1000000
  gasPerByte: 1
  gasPerSig: 1000
  gasFixed: 10000
configdbbadger:
  badgerDbPath: "./dbs/mpc_controller_db${id}"
EOM

# echo $config
echo -e "$CFG" > ./configs/config${id}.yaml
}

create_config 1
create_config 2
create_config 3

sleep 5

./mpc-controller --configFile ./configs/config1.yaml 2>&1 | tee logs/mpc-controller1.log &
./mpc-controller --configFile ./configs/config2.yaml 2>&1 | tee logs/mpc-controller2.log &
./mpc-controller --configFile ./configs/config3.yaml 2>&1 | tee logs/mpc-controller3.log &