#!/usr/bin/env bash

task tests:initiateStake

while true
do
	  bash ./tests/scripts/fund_initiateStake.sh
    sleep 10
    venom run tests/testsuites/initiateStake.yml
    sleep 30
done