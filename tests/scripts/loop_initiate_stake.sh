#!/usr/bin/env bash

LOOP_INITIATE_STAKE=0
MAX_LOOPS_ALLOWED=10000 # todo: enable command line config

echo Starting loop initiateStake request

while [ $LOOP_INITIATE_STAKE -lt $MAX_LOOPS_ALLOWED ]
do
#    bash ./tests/scripts/fund_participants.sh
	  bash ./tests/scripts/fund_initiateStake.sh

    venom run tests/testsuites/initiateStake.yml
    LOOP_INITIATE_STAKE=$((LOOP_INITIATE_STAKE+1))

    echo Looped initiateStake at $(date +%Y-%m-%d/%H:%M:%S), total times: $LOOP_INITIATE_STAKE
    sleep 26
done
