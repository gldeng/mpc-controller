#!/usr/bin/env bash

# Shared value
RPC_URL=http://127.0.0.1:9650/ext/bc/C/rpc

ROLE_DEFAULT_ADMIN="0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC" # Contract Deployer
ROLE_DEFAULT_ADMIN_PK="56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027" # Contract deployer PK

# Create temporary mpc-contract working directory
git submodule init
git submodule update

WD_CONTROLLER=$(pwd)

TMPDIR_CONTRACT=$(mktemp -d -t mpc-contract-$(date +%Y%m%d-%H%M%S)-XXX)
#export TMPDIR_CONTRACT=$TMPDIR_CONTRACT

cd $TMPDIR_CONTRACT

echo $TMPDIR_CONTRACT

forge init mpc-contract
cd mpc-contract

cp -a $WD_CONTROLLER/contract/src/. src/
cp -a $WD_CONTROLLER/contract/lib/. lib/

# Deploy MpcManager contract
MPC_MANAGER=$(forge create --rpc-url $RPC_URL --private-key $ROLE_DEFAULT_ADMIN_PK MpcManager | grep -i "deployed" | cut -d " " -f 3)
echo "MpcManager contract deployed to: "$MPC_MANAGER

# Deploy AvaLido contract
AVALIDO=$(forge create --rpc-url $RPC_URL --private-key $ROLE_DEFAULT_ADMIN_PK AvaLido --constructor-args  $MPC_MANAGER | grep -i "deployed" | cut -d " " -f 3)
echo "AvaLido contract deployed to: "$AVALIDO

# set AvaLido address for MpcManager
_=$(cast send --rpc-url $RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK $MPC_MANAGER "setAvaLidoAddress(address)" $AVALIDO)
echo "AvaLido set for MpcManager"

cd $WD_CONTROLLER
mkdir -p addresses
echo -n $MPC_MANAGER > addresses/MPC_MANAGER_ADDRESS
echo -n $AVALIDO > addresses/AVALIDO_ADDRESS

rm -rf $TMPDIR_CONTRACT