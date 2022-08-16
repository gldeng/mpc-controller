#!/usr/bin/env bash

# Contract deployer and its PK
ROLE_DEFAULT_ADMIN="0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"
ROLE_DEFAULT_ADMIN_PK="56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"

# Pause manager address
ROLE_PAUSE_MANAGER="0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"

# Addresses to receive principal and rewards after staking period ended
RECEIVE_PRINCIPAL_ADDR="0xd94fc5fd8812dde061f420d4146bc88e03b6787c"
RECEIVE_REWARD_ADDR="0xe8025f13e6bf0db21212b0dd6aebc4f3d1fb03ce"

# Network URLs
C_CHAIN_RPC_URL=http://127.0.0.1:9650/ext/bc/C/rpc

LAST_WD=$(pwd)

cd $HOME/mpctest/contracts/

echo "----------Start deploying contracts----------"

echo "Deploying ValidatorHelpers smart contract"
VALIDATOR_HELPERS=$(forge create --force --rpc-url $C_CHAIN_RPC_URL --private-key $ROLE_DEFAULT_ADMIN_PK ValidatorHelpers | grep -i "deployed" | cut -d " " -f 3)

echo "Deploying OracleManager smart contract"
ORACLE_MANAGER=$(forge create --force --rpc-url $C_CHAIN_RPC_URL --private-key $ROLE_DEFAULT_ADMIN_PK OracleManager --libraries src/Types.sol:ValidatorHelpers:$VALIDATOR_HELPERS  | grep -i "deployed" | cut -d " " -f 3)

echo "Deploying Oracle smart contract"
ORACLE=$(forge create --force --rpc-url $C_CHAIN_RPC_URL --private-key $ROLE_DEFAULT_ADMIN_PK Oracle | grep -i "deployed" | cut -d " " -f 3)

echo "Deploying ValidatorSelector smart contract"
VALIDATOR_SELECTOR=$(forge create --force --rpc-url $C_CHAIN_RPC_URL --private-key $ROLE_DEFAULT_ADMIN_PK ValidatorSelector --libraries src/Types.sol:ValidatorHelpers:$VALIDATOR_HELPERS  | grep -i "deployed" | cut -d " " -f 3)

echo "Deploying ParticipantIdHelpers smart contract"
PARTICIPANT_ID_HELPERS=$(forge create --force --rpc-url $C_CHAIN_RPC_URL --private-key $ROLE_DEFAULT_ADMIN_PK ParticipantIdHelpers | grep -i "deployed" | cut -d " " -f 3)

echo "Deploying ConfirmationHelpers smart contract"
CONFIRMATION_HELPERS=$(forge create --force --rpc-url $C_CHAIN_RPC_URL --private-key $ROLE_DEFAULT_ADMIN_PK ConfirmationHelpers | grep -i "deployed" | cut -d " " -f 3)

echo "Deploying KeygenStatusHelpers smart contract"
KEYGEN_STATUS_HELPERS=$(forge create --force --rpc-url $C_CHAIN_RPC_URL --private-key $ROLE_DEFAULT_ADMIN_PK KeygenStatusHelpers | grep -i "deployed" | cut -d " " -f 3)

echo "Deploying MpcManager smart contract"
MPC_MANAGER=$(forge create --force --rpc-url $C_CHAIN_RPC_URL --private-key $ROLE_DEFAULT_ADMIN_PK MpcManager --libraries src/MpcManager.sol:ParticipantIdHelpers:$PARTICIPANT_ID_HELPERS --libraries src/MpcManager.sol:ConfirmationHelpers:$CONFIRMATION_HELPERS --libraries src/MpcManager.sol:KeygenStatusHelpers:$KEYGEN_STATUS_HELPERS  | grep -i "deployed" | cut -d " " -f 3)

echo "Deploying AvaLido smart contract"
AVALIDO=$(forge create --force --rpc-url $C_CHAIN_RPC_URL --private-key $ROLE_DEFAULT_ADMIN_PK AvaLido --constructor-args  $MPC_MANAGER | grep -i "deployed" | cut -d " " -f 3)

echo "Initializing MpcManager contract"
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK $MPC_MANAGER "initialize(address,address,address,address,address)" $ROLE_DEFAULT_ADMIN $ROLE_PAUSE_MANAGER $AVALIDO $RECEIVE_PRINCIPAL_ADDR $RECEIVE_REWARD_ADDR > /dev/null

mkdir -p addresses
echo -n $MPC_MANAGER > addresses/MPC_MANAGER_ADDRESS
echo -n $AVALIDO > addresses/AVALIDO_ADDRESS

echo "----------Deployed contract addresses----------"

echo "ValidatorHelpers address:     $VALIDATOR_HELPERS"
echo "OracleManager address:        $ORACLE_MANAGER"
echo "Oracle address:               $ORACLE"
echo "ValidatorSelector address:    $VALIDATOR_SELECTOR"
echo "ParticipantIdHelpers address: $PARTICIPANT_ID_HELPERS"
echo "ConfirmationHelpers address:  $CONFIRMATION_HELPERS"
echo "KeygenStatusHelpers address:  $KEYGEN_STATUS_HELPERS"
echo "MpcManager address:           $MPC_MANAGER"
echo "AvaLido address:              $AVALIDO"

cd $LAST_WD
