package main

import (
	"flag"
	"log"
	"strconv"
	"strings"

	"github.com/panuwatphakaew/rafteeze/cmd"
)

func main() {
	id := flag.String("id", "1", "node ID")
	http := flag.String("http", ":8080", "HTTP server address for client API")
	grpc := flag.String("grpc", "", "gRPC server address for Raft transport (e.g., localhost:9001)")
	grpcPeersFlag := flag.String("grpc-peers", "", "comma-separated list of peer ID=gRPC address (e.g., 2=localhost:9002,3=localhost:9003)")
	flag.Parse()

	if *grpc == "" {
		log.Fatal("--grpc flag is required")
	}

	grpcPeers := make(map[uint64]string)
	if *grpcPeersFlag != "" {
		peerList := strings.Split(*grpcPeersFlag, ",")
		for _, peer := range peerList {
			parts := strings.Split(peer, "=")
			if len(parts) != 2 {
				log.Fatal("Invalid gRPC peer format. Expected ID=address")
			}
			peerID, err := strconv.ParseUint(parts[0], 10, 64)
			if err != nil {
				log.Fatal("Invalid peer ID:", err)
			}
			grpcPeers[peerID] = parts[1]
		}
	}

	if err := cmd.RunServer(*id, *http, *grpc, grpcPeers); err != nil {
		log.Fatal(err)
	}
}
