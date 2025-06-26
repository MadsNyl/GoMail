package main

import (
	"GoMail/logger"
	"GoMail/tasks"
	"log"
	"os"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// Initialize logger
	logger.InitLogger(
		os.Getenv("LOG_ENDPOINT"),
		100,
		"GoMail Worker",
	)

	logger.Info("Starting GoMail worker...", nil)

	redis_addr := os.Getenv("REDIS_ADDR")
	if redis_addr == "" {
		redis_addr = "localhost:6379"
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redis_addr},
		asynq.Config{Concurrency: 5},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeEmail, tasks.EmailTaskHandler)

	logger.Info("Worker running...", nil)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run worker: %v", err)
	}
}