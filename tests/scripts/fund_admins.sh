#!/usr/bin/env bash

# Contract deployer and its PK
ROLE_DEFAULT_ADMIN="0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"
ROLE_DEFAULT_ADMIN_PK="56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"

# Oracle admin
ROLE_ORACLE_ADMIN="0x8e7D0f159e992cfC0ee28D55C600106482a818Ea"

# Mpc admin
ROLE_MPC_ADMIN=$ROLE_ORACLE_ADMIN

# Network URLs
C_CHAIN_RPC_URL=http://127.0.0.1:9650/ext/bc/C/rpc

# fund the Oracle admin
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK --value 10ether $ROLE_ORACLE_ADMIN > /dev/null

# fund the Mpc admin
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK --value 10ether $ROLE_MPC_ADMIN > /dev/null
