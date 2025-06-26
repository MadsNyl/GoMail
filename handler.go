package main

import (
	"GoMail/logger"
	"GoMail/tasks"
	"encoding/json"
	"io"
	"net/http"

	"github.com/hibiken/asynq"
)

type EmailRequest struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}

func HandleEmail(w http.ResponseWriter, r *http.Request, client *asynq.Client) {
	if r.Method != http.MethodPost {
		logger.Error("Invalid request method", map[string]any{
			"method": r.Method,
			"endpoint": r.URL.Path,
		})
		http.Error(w, "Only POST supported", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed to read request body: " + err.Error(), map[string]any{
			"endpoint": r.URL.Path,
		})
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	var emailReq EmailRequest
	if err := json.Unmarshal(body, &emailReq); err != nil {
		logger.Error("Failed to parse JSON: " + err.Error(), map[string]any{
			"endpoint": r.URL.Path,
		})
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	task, err := tasks.NewEmailTask(emailReq.To, emailReq.Subject, emailReq.Body)

	logger.Info("Creating email task", map[string]any{
		"to":      emailReq.To,
		"subject": emailReq.Subject,
		"endpoint": r.URL.Path,
	})

	if err != nil {
		logger.Error("Failed to create email task: " + err.Error(), map[string]any{
			"endpoint": r.URL.Path,
		})
		http.Error(w, "Could not create task", http.StatusInternalServerError)
		return
	}

	info, err := client.Enqueue(task, asynq.MaxRetry(0))
	if err != nil {
		logger.Error("Failed to enqueue task: " + err.Error(), map[string]any{
			"to":      emailReq.To,
			"subject": emailReq.Subject,
			"endpoint": r.URL.Path,
		})
		http.Error(w, "Could not enqueue task", http.StatusInternalServerError)
		return
	}

	logger.Info("Email task enqueued", map[string]any{
		"task.id": info.ID,
		"queue":   info.Queue,
		"to":      emailReq.To,
		"subject": emailReq.Subject,
		"endpoint": r.URL.Path,
	})

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"task.id": info.ID,
		"queue":   info.Queue,
	}
	w.WriteHeader(http.StatusAccepted)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response: " + err.Error(), map[string]any{
			"endpoint": r.URL.Path,
		})
		http.Error(w, "Could not encode response", http.StatusInternalServerError)
		return
	}

	logger.Info("Email request handled successfully", map[string]any{
		"task.id": info.ID,
		"queue":   info.Queue,
		"to":      emailReq.To,
		"subject": emailReq.Subject,
		"endpoint": r.URL.Path,
	})
}