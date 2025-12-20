package main

import (
	"log"
	"os"
	"path/filepath"

	"mdu/cli"
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	logPath := filepath.Join(home, ".config", "mdu.log")

	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	log.SetOutput(logFile)
}

func main() {
	// Remove date/time prefix from log output
	log.SetFlags(0)

	if err := cli.NewRootCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}
