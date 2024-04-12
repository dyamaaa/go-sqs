package server

import (
	"encoding/json"
	"fmt"
	"go-sqs/internal/queue"
	"log"
	"net/http"
	"time"
)

// Constants for the enqueue and dequeue paths
const (
	EnqueuePath = "/enqueue"
	DequeuePath = "/dequeue"
)

// Server struct holds the queue manager
type Server struct {
	manager *queue.Manager
}

// NewServer creates a new server with the provided queue manager
func NewServer(manager *queue.Manager) *Server {
	return &Server{manager: manager}
}

// getQueueName extracts the queue name from the request
func (s *Server) getQueueName(r *http.Request) string {
	return r.URL.Query().Get("queueName")
}

// getQueue retrieves the queue with the given name, or creates a new one if it doesn't exist
func (s *Server) getQueue(queueName string, w http.ResponseWriter) (*queue.Queue, bool) {
	if queueName == "" {
		s.handleError(w, "Queue name is required", http.StatusBadRequest)
		return nil, false
	}

	q, err := s.manager.GetOrCreateQueue(queueName)
	if err != nil {
		s.handleError(w, fmt.Sprintf("Error getting or creating queue: %v", err), http.StatusInternalServerError)
		return nil, false
	}

	return q, true
}

// handleError logs the error and sends an HTTP response with the error message and status code
func (s *Server) handleError(w http.ResponseWriter, message string, statusCode int) {
	log.Printf("Error: %s, StatusCode: %d", message, statusCode)
	http.Error(w, fmt.Sprintf("{\"error\": \"%s\"}", message), statusCode)
}

// EnqueueHandler handles the enqueue requests
func (s *Server) EnqueueHandler(w http.ResponseWriter, r *http.Request) {
	queueName := s.getQueueName(r)
	q, ok := s.getQueue(queueName, w)
	if !ok {
		return
	}

	var msg queue.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("Error decoding message: %v", err)
		s.handleError(w, fmt.Sprintf("Error decoding message: %v", err), http.StatusBadRequest)
		return
	}

	if err := q.Enqueue(msg); err != nil {
		log.Printf("Error enqueuing message: %v", err)
		s.handleError(w, fmt.Sprintf("Error enqueuing message: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// DequeueHandler handles the dequeue requests
// DequeueHandler handles the HTTP request for dequeuing a message from the queue.
// It retrieves the queue name and timeout value from the request, gets the corresponding queue,
// dequeues a message with the specified timeout, and encodes the message as JSON in the response.
func (s *Server) DequeueHandler(w http.ResponseWriter, r *http.Request) {
	queueName := s.getQueueName(r)
	timeoutStr := r.URL.Query().Get("timeout")
	q, ok := s.getQueue(queueName, w)
	if !ok {
		return
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		log.Printf("Invalid timeout value %s: %v", timeoutStr, err)
		s.handleError(w, "Invalid timeout value", http.StatusBadRequest)
		return
	}

	msg, err := q.Dequeue(timeout)
	if err != nil {
		log.Printf("Error dequeuing message %s: %v", queueName, err)
		s.handleError(w, "Timeout or no message available", http.StatusRequestTimeout)
		return
	}

	if err := json.NewEncoder(w).Encode(msg); err != nil {
		log.Printf("Error encoding message: %v", err)
		s.handleError(w, fmt.Sprintf("Error encoding message: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) Run(addr string) error {
	http.HandleFunc(EnqueuePath, s.EnqueueHandler)
	http.HandleFunc(DequeuePath, s.DequeueHandler)
	log.Fatal(http.ListenAndServe(addr, nil))
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		return fmt.Errorf("server failed to start: %v", err)
	}
	return nil
}
