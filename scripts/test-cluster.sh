#!/usr/bin/env bash
set -e

trap "kill 0" EXIT

echo "Building rafteeze..."
go build -o rafteeze .

echo "Starting 3-node cluster..."

# Start node 1
./rafteeze -id 1 -http localhost:8001 -grpc localhost:9001 \
  -grpc-peers "2=localhost:9002,3=localhost:9003" &
sleep 1

# Start node 2
./rafteeze -id 2 -http localhost:8002 -grpc localhost:9002 \
  -grpc-peers "1=localhost:9001,3=localhost:9003" &
sleep 1

# Start node 3
./rafteeze -id 3 -http localhost:8003 -grpc localhost:9003 \
  -grpc-peers "1=localhost:9001,2=localhost:9002" &
sleep 2

echo ""
echo "Three-node cluster started!"
echo ""
echo "Test commands:"
echo "  # Write a key-value pair:"
echo "  curl -X PUT http://localhost:8001/kv/foo -d 'bar'"
echo ""
echo "  # Read from node 1:"
echo "  curl http://localhost:8001/kv/foo"
echo ""
echo "  # Read from node 2 (should be replicated):"
echo "  curl http://localhost:8002/kv/foo"
echo ""
echo "  # Read from node 3 (should be replicated):"
echo "  curl http://localhost:8003/kv/foo"
echo ""
echo "Press Ctrl+C to stop."

wait
