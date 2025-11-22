package main

import (
	"flag"
	"log"

	"github.com/panuwatphakaew/rafteeze/cmd"
)

func main() {
	id := flag.String("id", "1", "node ID")
	http := flag.String("http", ":8080", "http server address")
	raft := flag.String("raft", ":12000", "raft server address")
	join := flag.String("join", "", "existing node to join (for bootstrap)")
	flag.Parse()

	if err := cmd.RunServer(*id, *http, *raft, *join); err != nil {
		log.Fatal(err)
	}
}
