#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

NUM_PARTIES=3
PUBKEYS=("033217bb0e66dda25bcd50e2ccebabbe599312ae69c76076dd174e2fc5fdae73d8" "0272eab231c150b42e86cbe7398139432d2cad04289a820a922fe17b9d4ba577f4" "0373ee5cd601a19cd9bb95fe7be8b1566b73c51d3e7e375359c129b1d77bb4b3e6" "038196e06c3e803d0af06693a504ad14317550b4be4396ef57cf5f520c0f84833d" "03c20e0c088bb20027a77b1d23ad75058df5349c7a2bfafff7516c44c6f69aa66d" "03d0639e479fa1ca8ee13fd966c216e662408ff00349068bdc9c6966c4ea10fe3e" "02df7fb5bf5b3f97dffc98ecf8d660f604cad76f804a23e1b6cc76c11b5c92f345")
METRICPORTS=(":7001" ":7002" ":7003" ":7004" ":7005" ":7006" ":7007")

kill_process(){
  IND=$1
  DIR=tmp/party${IND}
  if test -f "$DIR/pid"; then
      echo "Killing MPC controller ${i}"
      kill -9 $(cat ${DIR}/pid)
  fi
}

start_mpc_controller(){
  IND=$1
  DIR=tmp/party${IND}
  mkdir -p $DIR
  echo "Starting MPC controller ${i}"
  ./mpc-controller \
        --host localhost \
        --port 9650 \
        --simulationMpcPrivateKey 56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027 \
        --metricsServeAddr ${METRICPORTS[$((IND-1))]} \
        --publicKey ${PUBKEYS[$((IND-1))]} \
        --dbPath ${DIR}/db \
        --keystoreDir /home/gldeng/local/avalido/internal-testnet/docker/mpc-controller/keystore/ \
        --passwordFile /home/gldeng/local/avalido/internal-testnet/docker/mpc-controller/password/password \
        --mpc-manager-address 0x549A3D41C0626ea686F8F208DE58761D8Dc61361 > ${DIR}/log.txt 2>&1 &
  echo $! > ${DIR}/pid
}

for i in $(seq $NUM_PARTIES);
  do kill_process $i;
done

echo "Starting MPC Controllers"
for i in $(seq $NUM_PARTIES);
  do start_mpc_controller $i;
done
