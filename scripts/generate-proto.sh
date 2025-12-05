#!/usr/bin/env bash
set -e

# Add Go bin to PATH for protoc plugins
export PATH="$HOME/go/bin:$PATH"

RAFT_VERSION="v3.5.25"
GOMODCACHE=$(go env GOMODCACHE)
RAFT_PATH="${GOMODCACHE}/go.etcd.io/etcd/raft/v3@${RAFT_VERSION}" # Path to etcd/raft module, not good, but works for now
GOGO_PATH="${GOMODCACHE}/github.com/gogo/protobuf@v1.3.2" # Path to gogo/protobuf module, not good, but works for now
### Note: Need more efficient way to manage proto dependencies ###

protoc --go_out=. --go_opt=paths=source_relative \
       --go_opt=Mraft.proto=go.etcd.io/etcd/raft/v3/raftpb \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       --go-grpc_opt=Mraft.proto=go.etcd.io/etcd/raft/v3/raftpb \
       -I. -I${RAFT_PATH}/raftpb -I${GOGO_PATH} \
       proto/transport.proto

echo "Proto files generated successfully"
