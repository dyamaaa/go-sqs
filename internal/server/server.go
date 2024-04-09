package server

import (
	"encoding/json"
	"fmt"
	"go-sqs/internal/queue"
	"log"
	"net/http"
	"time"
)

type Server struct {
	manager *queue.Manager
}

func NewServer(manager *queue.Manager) *Server {
	return &Server{manager: manager}
}

func (s *Server) getQueue(queueName string, w http.ResponseWriter) (*queue.Queue, bool) {
	if queueName == "" {
		http.Error(w, "Queue name is required", http.StatusBadRequest)
		return nil, false
	}

	q, err := s.manager.GetOrCreateQueue(queueName)
	if err != nil {
		s.handleError(w, fmt.Sprintf("Error getting or creating queue: %v", err), http.StatusInternalServerError)
		return nil, false
	}

	return q, true
}

func (s *Server) handleError(w http.ResponseWriter, message string, statusCode int) {
	log.Printf(message)
	http.Error(w, message, statusCode)
}

func (s *Server) EnqueueHandler(w http.ResponseWriter, r *http.Request) {
	queueName := r.URL.Query().Get("queueName")
	q, ok := s.getQueue(queueName, w)
	if !ok {
		return
	}

	var msg queue.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		s.handleError(w, fmt.Sprintf("Error decoding message: %v", err), http.StatusBadRequest)
		return
	}

	if err := q.Enqueue(msg); err != nil {
		s.handleError(w, fmt.Sprintf("Error enqueuing message: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) DequeueHandler(w http.ResponseWriter, r *http.Request) {
	queueName := r.URL.Query().Get("queueName")
	q, ok := s.getQueue(queueName, w)
	if !ok {
		return
	}

	msg, err := q.WaitForMessage(30 * time.Second)
	if err != nil {
		s.handleError(w, fmt.Sprintf("Error waiting for message: %v", err), http.StatusInternalServerError)
		return
	}

	if msg == nil {
		http.NotFound(w, r)
		return
	}

	if err := json.NewEncoder(w).Encode(msg); err != nil {
		s.handleError(w, fmt.Sprintf("Error encoding message: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) Run(addr string) {
	http.HandleFunc("/enqueue", s.EnqueueHandler)
	http.HandleFunc("/dequeue", s.DequeueHandler)

	log.Fatal(http.ListenAndServe(addr, nil))
}
