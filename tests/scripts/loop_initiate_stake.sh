#!/usr/bin/env bash

task tests:initiateStake

LOOP_INITIATE_STAKE=0

while true
do
	  bash ./tests/scripts/fund_initiateStake.sh
    sleep 30
    venom run tests/testsuites/initiateStake.yml
    LOOP_INITIATE_STAKE=$((LOOP_INITIATE_STAKE+1))
    echo loop initiateStake $LOOP_INITIATE_STAKE times
    sleep 300
done