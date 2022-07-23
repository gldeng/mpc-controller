#!/usr/bin/env bash

mkdir -p $HOME/mpctest/testwd

LAST_TEST_WD=$(mktemp -d -t mpctest-$(date +%Y%m%d%H%M%S)-XXX --tmpdir=$HOME/mpctest/testwd)

echo "Working directory: "$LAST_TEST_WD

echo $LAST_TEST_WD >  $HOME/mpctest/testwd_last

cd $LAST_TEST_WD

mkdir -p avalanchego
mkdir -p mpc-server
mkdir -p mpc-controller