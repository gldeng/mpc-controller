#!/usr/bin/env bash

LAST_WD=$(pwd)

echo "Start mpc-server-mock"

read LAST_TEST_WD < $HOME/mpctest/testwd_last

cd $LAST_TEST_WD/mpc-server-mock

mkdir log

$HOME/mpctest/mpc-server-mock/mpc-server-mock -p 7 - t 4 > log/log.log 2>&1 &

cd $LAST_WD