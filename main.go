package main

import (
	"GoMail/logger"
	"log"
	"net/http"
	"os"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main() {	
	// Load environment variables from .env file
	_ = godotenv.Load()

	// Initialize logger
	logger.InitLogger(
		os.Getenv("LOG_ENDPOINT"),
		100,
		"GoMail API",
	)

	logger.Info("Starting GoMail server...", nil)

	redisOpt, err := asynq.ParseRedisURI(os.Getenv("REDIS_ADDR"))
	if err != nil {
		logger.Error("Failed to parse Redis URI: " + err.Error(), nil)
		return
	}

	client := asynq.NewClient(redisOpt)
	defer client.Close()

	http.Handle("/", APIKeyAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleEmail(w, r, client)
	})))

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
