#!/usr/bin/env bash

pkill -f avalanchego
pkill -f mpc-controller
pkill -f mpc-server
pkill -f messenger
pkill -f oracle_service_mock

pkill -f ./tests/scripts/loop_initiate_stake.sh