package node

import (
	"encoding/json"

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
		ID:              1, //TODO: start multiple nodes with unique IDs
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   1024 * 1024, //TODO: Use configurable value
		MaxInflightMsgs: 256,         //TODO: Use configurable value
	}

	raftNode, err := raft.NewRawNode(cf, []raft.Peer{{ID: 1}})
	if err != nil {
		panic(err)
	}

	return &Node{
		id:   id,
		raft: raftNode,
		kv:   kv.NewInMemoryStore(),
	}
}

func (n *Node) Start() error {
	for {
		ready := n.raft.Ready()
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
