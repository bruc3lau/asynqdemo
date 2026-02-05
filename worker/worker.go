package worker

import (
	"asynqdemo/tasks"
	"log"

	"github.com/hibiken/asynq"
)

// StartWorker starts the Asynq worker server
func StartWorker(redisAddr string) error {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			// Number of concurrent workers
			Concurrency: 10,
		},
	)

	// Create a new ServeMux
	mux := asynq.NewServeMux()

	// Register task handlers
	mux.HandleFunc(tasks.TypeEmailDelivery, tasks.HandleEmailDeliveryTask)
	mux.HandleFunc(tasks.TypeDataProcess, tasks.HandleDataProcessTask)

	log.Println("🚀 Starting Asynq worker server...")
	log.Printf("📡 Connected to Redis at: %s", redisAddr)
	log.Println("👷 Worker is ready to process tasks")

	// Start the server (blocking call)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("Could not run server: %v", err)
		return err
	}

	return nil
}
