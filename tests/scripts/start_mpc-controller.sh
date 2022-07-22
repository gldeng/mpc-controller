#!/usr/bin/env bash

LAST_WD=$(pwd)

echo "Start mpc-controller"

read LAST_TEST_WD < $HOME/mpctest/testwd_last

cd $LAST_TEST_WD/mpc-controller

mkdir -p dbs
mkdir -p configs
mkdir -p logs

sks=("3db317d8f1ff081a32038c339901f8a6a15f53122dde3b99fa8017a2d0952f5ae5cdf5d0df912b8fc7755776a3b12af8cb7aece27b721d7f2c13b4cc957d1ff40131c7b18472748f2aeff2cf9a8aa46539d93281d367bdbe37179955824810ba" "49b5078e82a8ac2da6493fcc8da4ab8d97b3e2ca85e51b0d1fb5d271d4eaa5932a74784ce7672827d7f0ff4aeae1e73cadce9bf0ea75d6dc4eaab7ac8c81cb4a1c034f40d11940ffd65822a9d34f41830303516ac600e42b5e98ccf980efb67e" "f4c6214f7a5ec30236b9aaa2cddfc0963a4fafe52b29e2a2f0cf2c246de8fb11546d38591e2e45b9d9076bf6739282dce52b8f93651668875c784a1083adf4ae2af1253372ea921a1e5795975e372e829de2a1753e2d0bafcf760ef984881204" "83113ac0ec9e9afa9e5a4df6af5a07f3a5f6c6383ba0efecb70cc53c88b6278875839c83cf96af6f7ea7504d69f2196e1d23aa23c8f049b25508eb01fd1f240db97f3bce3f109eac0f012eb17c2009adbab83d5ca4ddf118699fc07ef6c74faf" "5f2f5f616b45a59e06f7ffe47c2b5207559edbdae11a596d5f8c06ed47d11e0399744461f96e7cff9adce8c0bea392d81e570df0da30b977487b24e5408ffcd349059517eb7d6626fc7e3e2483debf062a781dbed7a9a7cff5721c8d1ab8c13f" "e7700b12e822584c615fbf8f484cadd1f90b1b550461f9a9c9c53ca3680c99033b59c39b3478857b3e7ab886c6d93d577720c222203f0ed58034e70bf83858a39530f0d3a4f5a7de90b59d5685c4d920be9b51f7a2aeef18bb15db8a7866e5ef" "54bf73be16c9d97e59eb6b0d26f5290379b03b0cf20f3c1fd39cda9c64a048af8e73713142a29e5e6b4cd998edd63a7a79053943ce894794624723c80cc19992c4ffb0e8fdec96f7fd5036b5fc603c6a617c6851e0b8e42c86002e54f1e109f4")
MPC_SERVER_URLS=("http://localhost:8001" "http://localhost:8002" "http://localhost:8003" "http://localhost:8004" "http://localhost:8005" "http://localhost:8006" "http://localhost:8007")
#MPC_SERVER_URLS=("http://localhost:9000" "http://localhost:9000" "http://localhost:9000" "http://localhost:9000" "http://localhost:9000" "http://localhost:9000" "http://localhost:9000")


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
create_config 4
create_config 5
create_config 6
create_config 7

MPC_CONTROLLER_REPO=$HOME/mpctest/mpc-controller

$MPC_CONTROLLER_REPO/mpc-controller --configFile configs/config1.yaml --password k0n9MLBofTgo2DRVnxSM9hNw8GD9EZ8YTV3SZXwCNHqAtBzqgPJApCBLk0MvlJHt > logs/mpc-controller1.log 2>&1 &
$MPC_CONTROLLER_REPO/mpc-controller --configFile configs/config2.yaml --password pSCzMBSIKQXt2tOirE71vixMdobWJjhaCVdqm3IXJvwNRZ3r6r8So3IdEhWhPl1U > logs/mpc-controller2.log 2>&1 &
$MPC_CONTROLLER_REPO/mpc-controller --configFile configs/config3.yaml --password 0gXwUSG7PI4ylyKgL3WnPAF3qWQLnpy0jcu46ha9Fxc74RdylsOli4ZbfJ0e9CPg > logs/mpc-controller3.log 2>&1 &
$MPC_CONTROLLER_REPO/mpc-controller --configFile configs/config4.yaml --password kbxyZB5TF0x32qWLBeSUqMeTkQ4lO6otcvI6ZBbP0PcwO1t3vR52Cp6I8Pc1C25W > logs/mpc-controller4.log 2>&1 &
$MPC_CONTROLLER_REPO/mpc-controller --configFile configs/config5.yaml --password wIfYQEBMSXaPZ6coe8rYKXfZ1aE9jYj8FylK5W3c3tG8NgSsFCmIvWzk3EJA3Bly > logs/mpc-controller5.log 2>&1 &
$MPC_CONTROLLER_REPO/mpc-controller --configFile configs/config6.yaml --password fWGZCg4RtUwecc5amFZHcVhiSnqHa0mWVaUPeZH0gbp2B4f022Me0U3lDF7jG5jp > logs/mpc-controller6.log 2>&1 &
$MPC_CONTROLLER_REPO/mpc-controller --configFile configs/config7.yaml --password Pk4nRobphv2iE2A6qfx50Yo3DHBvzFCAbkedqsVUUKdGK7nBgKdyWod3buCC3qC2 > logs/mpc-controller7.log 2>&1 &

cd $LAST_WD
