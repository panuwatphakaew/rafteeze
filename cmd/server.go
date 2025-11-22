package cmd

import "github.com/panuwatphakaew/rafteeze/node"

func RunServer(id, httpAddr, raftAddr, joinAddr string) error {
	n := node.NewNode(id, httpAddr, raftAddr, joinAddr)
	return n.Start()
}