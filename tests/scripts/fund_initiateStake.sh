#!/usr/bin/env bash

# Contract deployer and its PK
ROLE_DEFAULT_ADMIN="0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"
ROLE_DEFAULT_ADMIN_PK="56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"

# Network URLs
C_CHAIN_RPC_URL=http://127.0.0.1:9650/ext/bc/C/rpc

MPC_MANAGER_ADDRESS=$(cat $HOME/mpctest/contracts/addresses/MPC_MANAGER_ADDRESS)

# Query LAST_GEN_ADDRESS
LAST_GEN_ADDRESS=$(cast call --rpc-url $C_CHAIN_RPC_URL $MPC_MANAGER_ADDRESS "lastGenAddress()")
LAST_GEN_ADDRESS=0x${LAST_GEN_ADDRESS: -40}

echo -n addr > $HOME/mpctest/contracts/addresses/LAST_GEN_ADDRESS

## Fund the LAST_GEN_ADDRESS so that they can afford gas fee
#cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK --value 1ether $LAST_GEN_ADDRESS > /dev/null

# Fund AvaLido address so that it have sufficient balance to initiate stake
# Note: 100000 ether for stake, 1 ether for gas
AVALIDO_ADDRESS=$(cat $HOME/mpctest/contracts/addresses/AVALIDO_ADDRESS)
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK --value 100001ether $AVALIDO_ADDRESS > /dev/null
