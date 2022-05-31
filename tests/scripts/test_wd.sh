#!/usr/bin/env bash

mkdir -p /tmp/mpctest/testwd

LAST_TEST_WD=$(mktemp -d -t mpctest-$(date +%Y%m%d%H%M%S)-XXX --tmpdir=/tmp/mpctest/testwd)

echo $LAST_TEST_WD >  /tmp/mpctest/testwd_last

cd $LAST_TEST_WD

mkdir -p avalanchego
mkdir -p mpc-server
mkdir -p mpc-controller