package cmd

import "github.com/panuwatphakaew/rafteeze/node"

func RunServer(id, httpAddr string) error {
	n := node.NewNode(id, httpAddr)
	return n.Start()
}
