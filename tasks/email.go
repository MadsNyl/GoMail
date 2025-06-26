package tasks

import (
	"GoMail/logger"
	"context"
	"encoding/json"
	"net/smtp"
	"os"

	"github.com/hibiken/asynq"
)

type EmailPayload struct {
	To      []string
	Subject string
	Body    string
}

func NewEmailTask(to []string, subject, body string) (*asynq.Task, error) {
	payload, err := json.Marshal(EmailPayload{
		To:      to,
		Subject: subject,
		Body:    body,
	})

	if err != nil {
		return nil, err
	}

	task := asynq.NewTask(TypeEmail, payload)

	return task, nil
}

func EmailTaskHandler(ctx context.Context, task *asynq.Task) error {
	var p EmailPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		return asynq.SkipRetry
	}

	auth := smtp.PlainAuth(
		"",
		os.Getenv("SMTP_USERNAME"),
		os.Getenv("SMTP_PASSWORD"),
		os.Getenv("SMTP_HOST"),
	)

	logger.Info("Sending email", map[string]any{
		"to":      p.To,
		"subject": p.Subject,
	})
	
	message := "Subject: " + p.Subject + "\n" + p.Body
	
	err := smtp.SendMail(
		os.Getenv("SMTP_ADDR"),
		auth,
		os.Getenv("SMTP_USERNAME"),
		p.To,
		[]byte(message),
	)

	if err != nil {
		message = "Failed to send email: " + err.Error()
		logger.Error(message, map[string]any{
			"to":      p.To,
			"subject": p.Subject,
		})
		return asynq.SkipRetry
	}

	logger.Info("Email sent successfully", map[string]any{
		"to":      p.To,
		"subject": p.Subject,
	})

	return nil
}
