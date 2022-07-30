#!/usr/bin/env bash

# Contract deployer and its PK
ROLE_DEFAULT_ADMIN="0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"
ROLE_DEFAULT_ADMIN_PK="56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"

# Addresses to receive principal and rewards after staking period ended
RECEIVE_PRINCIPAL_ADDR="0xd94fc5fd8812dde061f420d4146bc88e03b6787c"
RECEIVE_REWARD_ADDR="0xe8025f13e6bf0db21212b0dd6aebc4f3d1fb03ce"

# Network URLs
C_CHAIN_RPC_URL=http://127.0.0.1:9650/ext/bc/C/rpc

LAST_WD=$(pwd)

echo "Start deploying smart contracts"

#read LAST_TEST_WD < $HOME/mpctest/testwd_last
#
#cd $LAST_TEST_WD/contracts

cd $HOME/mpctest/contracts/

#rm -rf ./out
#rm -rf ./cache

# Deploy MpcManager contract
MPC_MANAGER=$(forge create --force --rpc-url $C_CHAIN_RPC_URL --private-key $ROLE_DEFAULT_ADMIN_PK MpcManager | grep -i "deployed" | cut -d " " -f 3)

# Deploy AvaLido contract
AVALIDO=$(forge create --force --rpc-url $C_CHAIN_RPC_URL --private-key $ROLE_DEFAULT_ADMIN_PK AvaLido --constructor-args  $MPC_MANAGER | grep -i "deployed" | cut -d " " -f 3)

# set AvaLido address for MpcManager
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK $MPC_MANAGER "setAvaLidoAddress(address)" $AVALIDO > /dev/null

# set address to receive principal
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK $MPC_MANAGER "setReceivePrincipalAddr(address)" $RECEIVE_PRINCIPAL_ADDR > /dev/null

# set address to receive rewards
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK $MPC_MANAGER "setReceiveRewardAddr(address)" $RECEIVE_REWARD_ADDR > /dev/null

mkdir -p addresses
echo -n $MPC_MANAGER > addresses/MPC_MANAGER_ADDRESS
echo -n $AVALIDO > addresses/AVALIDO_ADDRESS

echo "MpcManager address: $MPC_MANAGER"
echo "AvaLido address: $AVALIDO"

cd $LAST_WD