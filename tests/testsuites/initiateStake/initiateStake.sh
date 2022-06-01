#!/usr/bin/env bash

# todo: consider put these shared value in one place, to ensure single point of truth

# Contract deployer and its PK
ROLE_DEFAULT_ADMIN="0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"
ROLE_DEFAULT_ADMIN_PK="56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"

# Network URLs
C_CHAIN_RPC_URL=http://127.0.0.1:9650/ext/bc/C/rpc

# AvaLido address
AVALIDO_ADDRESS=$(cat /tmp/mpctest/contracts/addresses/AVALIDO_ADDRESS)

cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK --gas 900000 $AVALIDO_ADDRESS "initiateStake()"