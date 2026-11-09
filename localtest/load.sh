#!/usr/bin/env bash
# Sends N requests to each demo endpoint. Usage: localtest/load.sh [N]
N=${1:-500}
for i in $(seq 1 "$N"); do
  curl -s -o /dev/null http://127.0.0.1:18080/api/v1/greeting
  curl -s -o /dev/null "http://127.0.0.1:18080/api/v1/users/$i"
done
echo "sent $N requests to each endpoint"
