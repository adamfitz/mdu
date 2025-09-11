package main

import (
	"log"

	"mdu/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}
