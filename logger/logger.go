package logger

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type LogEntry struct {
	Timestamp         time.Time      `json:"timestamp"`
	SeverityText      string         `json:"severity_text"`
	SeverityNumber    int            `json:"severity_number"`
	Body              string         `json:"body"`
	ServiceName       *string        `json:"service_name,omitempty"`
	ServiceVersion    *string        `json:"service_version,omitempty"`
	HostName          *string        `json:"host_name,omitempty"`
	Attributes        map[string]any `json:"attributes,omitempty"`
}

var (
	logQueue     chan LogEntry
	once         sync.Once
	logEndpoint  string
	defaultSvc   string
	defaultHost  string
	defaultVersion = "1.0.0"
)

// InitLogger should be called once at app startup.
func InitLogger(endpoint string, bufferSize int, serviceName string) {
	once.Do(func() {
		logEndpoint = endpoint
		defaultSvc = serviceName
		defaultHost, _ = os.Hostname()
		logQueue = make(chan LogEntry, bufferSize)
		go worker()
	})
}

// Log is the internal function to enqueue a log entry.
func Log(severity string, severityNum int, body string, attributes map[string]any) {
	now := time.Now().UTC()

	entry := LogEntry{
		Timestamp:      now,
		SeverityText:   severity,
		SeverityNumber: severityNum,
		Body:           body,
		ServiceName:    &defaultSvc,
		HostName:       &defaultHost,
		ServiceVersion: &defaultVersion,
		Attributes:     attributes,
	}

	select {
		case logQueue <- entry:
		default:
			log.Println("logger: log queue full, dropping entry")
	}
}

// worker continuously sends logs in the background.
func worker() {
	for entry := range logQueue {
		go send(entry)
	}
}

// send performs the HTTP POST to the central log endpoint.
func send(entry LogEntry) {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		log.Printf("logger: failed to marshal log entry: %v", err)
		return
	}

	username := os.Getenv("LOG_USERNAME")
	password := os.Getenv("LOG_PASSWORD")

	req, err := http.NewRequest("POST", logEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("logger: failed to create request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("logger: failed to send log: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("logger: log server responded with status: %v", resp.Status)
	}
}

func Info(body string, attributes map[string]any) {
	Log("INFO", 9, body, attributes)
}

func Warning(body string, attributes map[string]any) {
	Log("WARNING", 13, body, attributes)
}

func Error(body string, attributes map[string]any) {
	Log("ERROR", 17, body, attributes)
}

func Debug(body string, attributes map[string]any) {
	Log("DEBUG", 5, body, attributes)
}
