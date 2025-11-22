#!/usr/bin/env bash
set -e

trap "kill 0" EXIT

go build -o tmp/rafteeze .

/tmp/rafteeze --id node1 --http 8001 --raft 9002 &
sleep 1
/tmp/rafteeze --id node2 --http 8002 --raft 9003 --join 9002 &
/tmp/rafteeze --id node3 --http 8003 --raft 9004 --join 9002 &

echo "Three-node cluster started. Press Ctrl+C to stop."
echo "Try: curl -XPOST localhost:8001/kv/foo -d 'bar'"
wait