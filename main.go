package main

import (
	"flag"
	"log"

	"github.com/panuwatphakaew/rafteeze/cmd"
)

func main() {
	id := flag.String("id", "1", "node ID")
	http := flag.String("http", ":8080", "http server address")
	flag.Parse()

	if err := cmd.RunServer(*id, *http); err != nil {
		log.Fatal(err)
	}
}
