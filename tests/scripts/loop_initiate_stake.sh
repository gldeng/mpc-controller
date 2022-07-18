#!/usr/bin/env bash

task tests:initiateStake

while true
do
	  bash ./tests/scripts/fund_initiateStake.sh
    sleep 30
    venom run tests/testsuites/initiateStake.yml
    sleep 60
done