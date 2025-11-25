package node

import (
	"encoding/json"
	"log"
	"time"

	"github.com/coreos/etcd/raft/raftpb"
	"github.com/panuwatphakaew/rafteeze/kv"
	"go.etcd.io/etcd/raft"
)

type Node struct {
	id   string
	raft *raft.RawNode
	kv   *kv.Store
}

func NewNode(id, httpAddr string) *Node {
	storage := raft.NewMemoryStorage()

	cf := &raft.Config{
		ID:              1, //TODO: Start multiple nodes with unique IDs
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   1024 * 1024, //TODO: Use configurable value
		MaxInflightMsgs: 256,         //TODO: Use configurable value
	}

	raftNode, err := raft.NewRawNode(cf, []raft.Peer{{ID: 1}}) //Hardcoded single node for now
	if err != nil {
		log.Fatal("Failed to create Raft node:", err)
	}

	return &Node{
		id:   id,
		raft: raftNode,
		kv:   kv.NewInMemoryStore(),
	}
}

func (n *Node) Start() error {
	for {
		if n.raft.HasReady() {
			ready := n.raft.Ready() //TODO: Optimize Ready handling with channels and goroutines

			if len(ready.Entries) > 0 {
				// TODO: Persist ready.Entries to disk or other stable storage
			}

			if len(ready.Messages) > 0 {
				// TODO: Send ready.Messages to peers via transport
			}

			for _, entry := range ready.CommittedEntries {
				if entry.Type == raftpb.EntryNormal && len(entry.Data) > 0 {
					payload := kv.Payload{}
					err := json.Unmarshal(entry.Data, &payload)
					if err != nil {
						return err
					}
					n.kv.Put(payload.Key, payload.Value)
				}
			}
			n.raft.Advance(ready)
		} else {
			time.Sleep(10 * time.Millisecond)
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
	return n.raft.Propose(data)
}
