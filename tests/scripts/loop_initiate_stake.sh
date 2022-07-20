#!/usr/bin/env bash

task tests:initiateStake

LOOP_INITIATE_STAKE=0

echo Starting loop initiateStake request

while true
do
	  bash ./tests/scripts/fund_initiateStake.sh
    sleep 30
    venom run tests/testsuites/initiateStake.yml
    LOOP_INITIATE_STAKE=$((LOOP_INITIATE_STAKE+1))
    echo Looped initiateStake request times: $LOOP_INITIATE_STAKE
    sleep 300
done