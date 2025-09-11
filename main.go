package main

import (
	"log"

	"mdu/cli"
)

func main() {
	// Remove date/time prefix from log output
	log.SetFlags(0)

	if err := cli.NewRootCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}
