package node

import (
	"github.com/panuwatphakaew/rafteeze/kv"
	"github.com/panuwatphakaew/rafteeze/raft"
)

type Node struct {
	id   string
	raft *raft.Raft
	kv   *kv.Store
}

func NewNode(id, httpAddr, raftAddr, joinAddr string) *Node {
	return &Node{
		id: id,
		kv: kv.NewInMemoryStore(),
	}
}

func (n *Node) Start() error {
	// Initialize Raft and KV store here
	select {}
}
