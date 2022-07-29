#!/usr/bin/env bash

# Contract deployer and its PK
ROLE_DEFAULT_ADMIN="0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"
ROLE_DEFAULT_ADMIN_PK="56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"

# Network URLs
C_CHAIN_RPC_URL=http://127.0.0.1:9650/ext/bc/C/rpc

# fund the participants so that they can afford gas fee
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK --value 10ether 0x3051bA2d313840932B7091D2e8684672496E9A4B > /dev/null
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK --value 10ether 0x7Ac8e2083E3503bE631a0557b3f2A8543EaAdd90 > /dev/null
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK --value 10ether 0x3600323b486F115CE127758ed84F26977628EeaA > /dev/null
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK --value 10ether 0xb418a71fbb8241116971b935ca8a77246b4B4389 > /dev/null
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK --value 10ether 0xC9e736Bda0d75ce0A8CBA0500c9f6D3EE6fE317a > /dev/null
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK --value 10ether 0x556fBDACA802fdaB2d6FC6173fFb52318834A0e5 > /dev/null
cast send --rpc-url $C_CHAIN_RPC_URL --from $ROLE_DEFAULT_ADMIN --private-key $ROLE_DEFAULT_ADMIN_PK --value 10ether 0x2E40efEEF4d95139c74BCB117c1C0Ba4B8Ac53db > /dev/null