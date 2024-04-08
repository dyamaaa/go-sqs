package main

import (
	"go-sqs/internal/queue"
	"go-sqs/internal/server"
	"log"
)

func main() {
	dbPath := "./tmp/badger"
	q, err := queue.NewManager(dbPath)
	if err != nil {
		log.Fatalf("Failed to create queue: %v", err)
	}

	srv := server.NewServer(q)
	srv.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
