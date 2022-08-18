#!/usr/bin/env bash

# Contract deployer and its PK
ROLE_DEFAULT_ADMIN="0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"
ROLE_DEFAULT_ADMIN_PK="56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"

# Network URLs
C_CHAIN_RPC_URL=http://127.0.0.1:9650/ext/bc/C/rpc

LAST_WD=$(pwd)

cd $HOME/mpctest/contracts/

#forge script src/deploy/Deploy.t.sol --sig "deploy()" --broadcast --rpc-url $C_CHAIN_RPC_URL --private-key $ROLE_DEFAULT_ADMIN_PK

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

#echo "----------Deployed contract addresses----------"

echo "AvaLido address:              $AVALIDO"
echo "ValidatorSelector address:    $VALIDATOR_SELECTOR"
echo "Oracle address:               $ORACLE"
echo "OracleManager address:        $ORACLE_MANAGER"
echo "MpcManager address:           $MPC_MANAGER"

cd $LAST_WD
