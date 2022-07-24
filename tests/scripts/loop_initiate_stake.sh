#!/usr/bin/env bash

LOOP_INITIATE_STAKE=0

echo Starting loop initiateStake request

while true
do
    bash ./tests/scripts/fund_participants.sh
    sleep 5
	  bash ./tests/scripts/fund_initiateStake.sh
    sleep 5
    venom run tests/testsuites/initiateStake.yml
    LOOP_INITIATE_STAKE=$((LOOP_INITIATE_STAKE+1))
    echo Looped initiateStake at $(date +%Y-%m-%d/%H:%M:%S), total times: $LOOP_INITIATE_STAKE
    sleep 5
done