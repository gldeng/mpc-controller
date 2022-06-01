#!/usr/bin/env bash

# todo: consider put these shared value in one place, to ensure single point of truth

# Contract deployer and its PK
ROLE_DEFAULT_ADMIN="0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"
ROLE_DEFAULT_ADMIN_PK="56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"

# Network URLs
C_CHAIN_RPC_URL=http://127.0.0.1:9650/ext/bc/C/rpc

# MpcManager address
MPC_MANAGER_ADDRESS=$(cat /tmp/mpctest/contracts/addresses/MPC_MANAGER_ADDRESS)

# Pre-defined group ID
MPC_GROUP_ID="3726383e52fd4cb603498459e8a4a15d148566a51b3f5bfbbf3cac7b61647d04"

# Request keygen
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK $MPC_MANAGER_ADDRESS "requestKeygen(bytes32)" $MPC_GROUP_ID

# CHeck keygen result
