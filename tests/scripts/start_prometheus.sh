#!/usr/bin/env bash

promImage=$(docker image ls | grep prom/prometheus)
if [ -z "$promImage" ]; then
  echo "Pulling prom/prometheus"
  docker pull prom/prometheus
fi

docker run --network="host"  -p 9090:9090 -v $(pwd)/tests/configs/prom:/etc/prometheus prom/prometheus > /dev/null