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

	redisOpt, err := asynq.ParseRedisURI(os.Getenv("REDIS_ADDR"))
	if err != nil {
		logger.Error("Failed to parse Redis URI: " + err.Error(), nil)
		return
	}

	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{Concurrency: 5},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeEmail, tasks.EmailTaskHandler)

	logger.Info("Worker running...", nil)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run worker: %v", err)
	}
}