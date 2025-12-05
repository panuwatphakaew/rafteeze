package node

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"strconv"
	"time"

	httpserver "github.com/panuwatphakaew/rafteeze/http"
	"github.com/panuwatphakaew/rafteeze/kv"
	"github.com/panuwatphakaew/rafteeze/transport"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type Node struct {
	id        uint64
	httpAddr  string
	grpcAddr  string
	raft      raft.Node
	storage   *raft.MemoryStorage
	kv        *kv.Store
	grpcPeers map[uint64]string // peer ID -> gRPC address
	transport *transport.Transport
}

func NewNode(id, httpAddr, grpcAddr string, grpcPeers map[uint64]string) (*Node, error) {
	nodeID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		log.Fatal("Invalid node ID:", err)
	}

	storage := raft.NewMemoryStorage()

	var peerIDs []uint64
	for peerID := range grpcPeers {
		peerIDs = append(peerIDs, peerID)
	}
	peerIDs = append(peerIDs, nodeID)

	confState := raftpb.ConfState{Voters: peerIDs}
	snapshot := raftpb.Snapshot{
		Metadata: raftpb.SnapshotMetadata{
			ConfState: confState,
			Index:     1,
			Term:      1,
		},
	}
	if err := storage.ApplySnapshot(snapshot); err != nil {
		log.Fatal("Failed to apply snapshot:", err)
	}

	cf := &raft.Config{
		ID:              nodeID,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		Applied:         1,
		MaxSizePerMsg:   1024 * 1024, //TODO: Use configurable value
		MaxInflightMsgs: 256,         //TODO: Use configurable value
	}

	raftNode := raft.RestartNode(cf)

	// Create gRPC transport
	tr, err := transport.NewTransport(nodeID, grpcAddr, grpcPeers)
	if err != nil {
		log.Fatal("Failed to create transport:", err)
	}

	return &Node{
		id:        nodeID,
		httpAddr:  httpAddr,
		grpcAddr:  grpcAddr,
		raft:      raftNode,
		storage:   storage,
		kv:        kv.NewInMemoryStore(),
		grpcPeers: grpcPeers,
		transport: tr,
	}, nil
}

func (n *Node) Start() error {
	// Start HTTP server for client API
	server := httpserver.NewServer(n.httpAddr, n.Propose, n.Get)
	go func() {
		if err := server.Start(); err != nil {
			log.Fatal("HTTP server failed:", err)
		}
	}()

	// Start gRPC server for Raft transport
	go func() {
		if err := n.transport.Start(); err != nil {
			log.Fatal("gRPC transport failed:", err)
		}
	}()

	slog.Info("Node started",
		slog.Uint64("id", n.id),
		slog.String("http", n.httpAddr),
		slog.String("grpc", n.grpcAddr))

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	//TODO: Use signal handling to gracefully shutdown the node
	for {
		select {
		case <-ticker.C:
			n.raft.Tick()

		case msg := <-n.transport.ReceiveMessage():
			if err := n.raft.Step(context.TODO(), msg); err != nil {
				slog.Error("failed to step message", slog.Any("error", err))
			}

		case ready := <-n.raft.Ready():
			if !raft.IsEmptySnap(ready.Snapshot) {
				if err := n.storage.ApplySnapshot(ready.Snapshot); err != nil {
					slog.Error("failed to apply snapshot", slog.Any("error", err))
				}
			}

			if len(ready.Entries) > 0 {
				if err := n.storage.Append(ready.Entries); err != nil {
					slog.Error("failed to append entries", slog.Any("error", err))
				}
			}

			if !raft.IsEmptyHardState(ready.HardState) {
				if err := n.storage.SetHardState(ready.HardState); err != nil {
					slog.Error("failed to set hard state", slog.Any("error", err))
				}
			}

			for _, msg := range ready.Messages {
				go func(m raftpb.Message) {
					if err := n.transport.Send(m); err != nil {
						slog.Error("failed to send message", slog.Any("error", err))
					}
				}(msg)
			}

			for _, entry := range ready.CommittedEntries {
				if entry.Type == raftpb.EntryNormal && len(entry.Data) > 0 {
					payload := kv.Payload{}
					err := json.Unmarshal(entry.Data, &payload)
					if err != nil {
						slog.Error("error process entry", slog.Any("error", err))
					} else {
						n.kv.Put(payload.Key, payload.Value)
						slog.Info("applied entry", slog.String("key", payload.Key), slog.String("value", payload.Value))
					}
				}
			}

			n.raft.Advance()
		}
	}
}

func (n *Node) Propose(key, value string) error {
	payload := kv.Payload{
		Key:   key,
		Value: value,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return n.raft.Propose(context.TODO(), data)
}

func (n *Node) Get(key string) (string, bool) {
	return n.kv.Get(key)
}
