#!/usr/bin/env bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/config.sh

# Setup Oracle address for Oracle Manager
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_ORACLE_ADMIN --private-key $ROLE_ORACLE_ADMIN_PK --gas-limit 900000 $ORACLE_MANAGER_ADDRESS "setOracleAddress(address)" $ORACLE_ADDRESS > /dev/null

# Set node ID list for Oracle
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_ORACLE_ADMIN --private-key $ROLE_ORACLE_ADMIN_PK --gas-limit 900000 $ORACLE_ADDRESS "setNodeIDList(string[])" $NODE_ID_LIST > /dev/null

# Set epoch duration for Oracle
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_ORACLE_ADMIN --private-key $ROLE_ORACLE_ADMIN_PK --gas-limit 900000 $ORACLE_ADDRESS "setEpochDuration(uint256)" 17 > /dev/null

# Set max protocol controlled AVAX for AvaLido
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_PROTOCOL_MANAGER --private-key $ROLE_PROTOCOL_MANAGER_PK --gas-limit 900000 $AVALIDO_ADDRESS "setMaxProtocolControlledAVAX(uint256)" 50000000000000000000000000 > /dev/null

# Set stake period for AvaLido
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_PROTOCOL_MANAGER --private-key $ROLE_PROTOCOL_MANAGER_PK --gas-limit 900000 $AVALIDO_ADDRESS "setStakePeriod(uint256)" 600 > /dev/null

# Set P-Chain export buffer for AvaLido
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_PROTOCOL_MANAGER --private-key $ROLE_PROTOCOL_MANAGER_PK --gas-limit 900000 $AVALIDO_ADDRESS "setPChainExportBuffer(uint256)" 300 > /dev/null
