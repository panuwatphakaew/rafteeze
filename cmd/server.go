package cmd

import (
	"github.com/panuwatphakaew/rafteeze/node"
)

func RunServer(id, httpAddr, grpcAddr string, grpcPeers map[uint64]string) error {
	n, err := node.NewNode(id, httpAddr, grpcAddr, grpcPeers)
	if err != nil {
		return err
	}
	return n.Start()
}
