#!/usr/bin/env bash

# Contract deployer and its PK
ROLE_DEFAULT_ADMIN="0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"
ROLE_DEFAULT_ADMIN_PK="56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"

# Network URLs
C_CHAIN_RPC_URL=http://127.0.0.1:9650/ext/bc/C/rpc

LAST_WD=$(pwd)

echo "Starting deploy smart contracts"

#read LAST_TEST_WD < /tmp/mpctest/testwd_last
#
#cd $LAST_TEST_WD/contracts

cd /tmp/mpctest/contracts/

# Deploy MpcManager contract
MPC_MANAGER=$(forge create --rpc-url $C_CHAIN_RPC_URL --private-key $ROLE_DEFAULT_ADMIN_PK MpcManager | grep -i "deployed" | cut -d " " -f 3)
echo "MpcManager contract deployed to: "$MPC_MANAGER

# Deploy AvaLido contract
AVALIDO=$(forge create --rpc-url $C_CHAIN_RPC_URL --private-key $ROLE_DEFAULT_ADMIN_PK AvaLido --constructor-args  $MPC_MANAGER | grep -i "deployed" | cut -d " " -f 3)
echo "AvaLido contract deployed to: "$AVALIDO

# set AvaLido address for MpcManager
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK $MPC_MANAGER "setAvaLidoAddress(address)" $AVALIDO > /dev/null
echo "AvaLido set for MpcManager"

mkdir -p addresses
echo -n $MPC_MANAGER > addresses/MPC_MANAGER_ADDRESS
echo -n $AVALIDO > addresses/AVALIDO_ADDRESS

cd $LAST_WD