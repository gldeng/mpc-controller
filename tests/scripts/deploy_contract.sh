#!/usr/bin/env bash

# Contract deployer and its PK
ROLE_DEFAULT_ADMIN="0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"
ROLE_DEFAULT_ADMIN_PK="56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"

# Oracle admin and its PK
ROLE_ORACLE_ADMIN="0x8e7D0f159e992cfC0ee28D55C600106482a818Ea"
ROLE_ORACLE_ADMIN_PK="a87518b3691061b9de6dd281d2dda06a4fe3a2c1b4621ac1e05d9026f73065bd"

# Protocol manager and its PK
ROLE_PROTOCOL_MANAGER=$ROLE_DEFAULT_ADMIN
ROLE_PROTOCOL_MANAGER_PK=$ROLE_DEFAULT_ADMIN_PK

# Network URLs
C_CHAIN_RPC_URL=http://127.0.0.1:9650/ext/bc/C/rpc

# Node ID list
NODE_ID_LIST=["NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5","NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5","NodeID-NFBbbJ4qCmNaCzeW7sxErhvWqvEQMnYcN","NodeID-MFrZFVCXPv5iCn6M9K6XduxGTYp891xXZ","NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg"]

LAST_WD=$(pwd)

cd $HOME/mpctest/contracts/

CONTRACTS=$(forge script src/deploy/Deploy.t.sol --sig "deploy()" --broadcast --rpc-url $C_CHAIN_RPC_URL --private-key $ROLE_DEFAULT_ADMIN_PK | grep Deployed)
# todo: simplify RE
AVALIDO=$(echo $CONTRACTS  | grep -o 'Deployed AvaLido, \w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w' | grep -o '0x\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w')
VALIDATOR_SELECTOR=$(echo $CONTRACTS  | grep -o 'Deployed Validator Selector, \w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w' | grep -o '0x\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w')
ORACLE=$(echo $CONTRACTS  | grep -o 'Deployed Oracle, \w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w' | grep -o '0x\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w')
ORACLE_MANAGER=$(echo $CONTRACTS  | grep -o 'Deployed Oracle Manager, \w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w' | grep -o '0x\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w')
MPC_MANAGER=$(echo $CONTRACTS  | grep -o 'Deployed MPC Manager, \w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w' | grep -o '0x\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w\w')

mkdir -p addresses
echo -n $MPC_MANAGER > addresses/MPC_MANAGER_ADDRESS
echo -n $AVALIDO > addresses/AVALIDO_ADDRESS
echo -n $ORACLE_MANAGER > addresses/ORACLE_MANAGER_ADDRESS

#echo "----------Deployed contract addresses----------"

echo "AvaLido address:              $AVALIDO"
echo "ValidatorSelector address:    $VALIDATOR_SELECTOR"
echo "Oracle address:               $ORACLE"
echo "OracleManager address:        $ORACLE_MANAGER"
echo "MpcManager address:           $MPC_MANAGER"

# Setup Oracle address for Oracle Manager
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_ORACLE_ADMIN --private-key $ROLE_ORACLE_ADMIN_PK --gas-limit 900000 $ORACLE_MANAGER "setOracleAddress(address)" $ORACLE > /dev/null

# Set node ID list for Oracle
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_ORACLE_ADMIN --private-key $ROLE_ORACLE_ADMIN_PK --gas-limit 900000 $ORACLE "setNodeIDList(string[])" $NODE_ID_LIST > /dev/null

# Set stake period for AvaLido
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_PROTOCOL_MANAGER --private-key $ROLE_PROTOCOL_MANAGER_PK --gas-limit 900000 $AVALIDO "setStakePeriod(uint256)" 600 > /dev/null

# Set P-Chain export buffer
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_PROTOCOL_MANAGER --private-key $ROLE_PROTOCOL_MANAGER_PK --gas-limit 900000 $AVALIDO "setPChainExportBuffer(uint256)" 300 > /dev/null

cd $LAST_WD