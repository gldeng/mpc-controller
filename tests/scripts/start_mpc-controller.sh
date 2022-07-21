#!/usr/bin/env bash

LAST_WD=$(pwd)

echo "Start mpc-controller"

read LAST_TEST_WD < $HOME/mpctest/testwd_last

cd $LAST_TEST_WD/mpc-controller

mkdir -p dbs
mkdir -p configs
mkdir -p logs

sks=("3db317d8f1ff081a32038c339901f8a6a15f53122dde3b99fa8017a2d0952f5ae5cdf5d0df912b8fc7755776a3b12af8cb7aece27b721d7f2c13b4cc957d1ff40131c7b18472748f2aeff2cf9a8aa46539d93281d367bdbe37179955824810ba" "49b5078e82a8ac2da6493fcc8da4ab8d97b3e2ca85e51b0d1fb5d271d4eaa5932a74784ce7672827d7f0ff4aeae1e73cadce9bf0ea75d6dc4eaab7ac8c81cb4a1c034f40d11940ffd65822a9d34f41830303516ac600e42b5e98ccf980efb67e" "f4c6214f7a5ec30236b9aaa2cddfc0963a4fafe52b29e2a2f0cf2c246de8fb11546d38591e2e45b9d9076bf6739282dce52b8f93651668875c784a1083adf4ae2af1253372ea921a1e5795975e372e829de2a1753e2d0bafcf760ef984881204")
MPC_SERVER_URLS=("http://localhost:8001" "http://localhost:8002" "http://localhost:8003")
#MPC_SERVER_URLS=("http://localhost:9000" "http://localhost:9000" "http://localhost:9000")


MPC_MANAGER_ADDRESS=$(cat $HOME/mpctest/contracts/addresses/MPC_MANAGER_ADDRESS)
function create_config(){
    id=$1
    sk=${sks[$(expr ${id} - 1)]}
    mpcServerUrl=${MPC_SERVER_URLS[$(expr ${id} - 1)]}
    read -r -d '' CFG <<- EOM
enableDevMode: true
controllerId: "mpc-controller-0${id}"
controllerKey: "${sk}"
mpcManagerAddress: "${MPC_MANAGER_ADDRESS}"
mpcServerUrl: "${mpcServerUrl}"
ethRpcUrl: "http://localhost:9650/ext/bc/C/rpc"
ethWsUrl: "ws://127.0.0.1:9650/ext/bc/C/ws"
cChainIssueUrl: "http://localhost:9650"
pChainIssueUrl: "http://localhost:9650"
networkConfig:
  networkId: 12345
  chainId: 43112
  cChainId: "2CA6j5zYzasynPsFeNoqWkmTCt3VScMvXUZHbfDJ8k3oGzAPtU"
  avaxId: "2fombhL7aGPwj3KH4bfrmJwW6PVnMobf9Y2fn9GwxiAAJyFDbe"
  importFee: 1000000
  exportFee: 1000000
  gasPerByte: 1
  gasPerSig: 1000
  gasFixed: 10000
databaseConfig:
  badgerDbPath: "./dbs/mpc_controller_db${id}"
EOM

echo -e "$CFG" > ./configs/config${id}.yaml
}

create_config 1
create_config 2
create_config 3

MPC_CONTROLLER_REPO=$HOME/mpctest/mpc-controller

$MPC_CONTROLLER_REPO/mpc-controller --configFile configs/config1.yaml --password k0n9MLBofTgo2DRVnxSM9hNw8GD9EZ8YTV3SZXwCNHqAtBzqgPJApCBLk0MvlJHt > logs/mpc-controller1.log 2>&1 &
$MPC_CONTROLLER_REPO/mpc-controller --configFile configs/config2.yaml --password pSCzMBSIKQXt2tOirE71vixMdobWJjhaCVdqm3IXJvwNRZ3r6r8So3IdEhWhPl1U > logs/mpc-controller2.log 2>&1 &
$MPC_CONTROLLER_REPO/mpc-controller --configFile configs/config3.yaml --password 0gXwUSG7PI4ylyKgL3WnPAF3qWQLnpy0jcu46ha9Fxc74RdylsOli4ZbfJ0e9CPg > logs/mpc-controller3.log 2>&1 &

cd $LAST_WD
