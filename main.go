package main

import (
	"asynqdemo/server"
	"asynqdemo/worker"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	redisAddr  = "localhost:6379"
	serverAddr = ":3000"
)

func main() {
	log.Println("🚀 Starting Asynq Demo Application")
	log.Println("=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=")

	// Start worker in a separate goroutine
	go func() {
		if err := worker.StartWorker(redisAddr); err != nil {
			log.Fatalf("Failed to start worker: %v", err)
		}
	}()

	// Start HTTP server
	srv := server.NewServer(redisAddr)
	defer srv.Close()

	// Handle graceful shutdown
	go func() {
		if err := srv.Run(serverAddr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")
}
