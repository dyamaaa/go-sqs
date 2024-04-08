package server

import (
	"encoding/json"
	"go-sqs/internal/queue"
	"log"
	"net/http"
)

type Server struct {
	manager *queue.Manager
}

func NewServer(manager *queue.Manager) *Server {
	return &Server{manager: manager}
}

func (s *Server) EnqueueHandler(w http.ResponseWriter, r *http.Request) {
	queueName := r.URL.Query().Get("queueName")
	if queueName == "" {
		http.Error(w, "Queue name is required", http.StatusBadRequest)
		return
	}

	var msg queue.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	q, err := s.manager.GetOrCreateQueue(queueName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := q.Enqueue(msg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) DequeueHandler(w http.ResponseWriter, r *http.Request) {
	queueName := r.URL.Query().Get("queueName")
	if queueName == "" {
		http.Error(w, "Queue name is required", http.StatusBadRequest)
		return
	}

	q, err := s.manager.GetOrCreateQueue(queueName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	msg, err := q.Dequeue()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if msg == nil {
		http.NotFound(w, r)
		return
	}

	if err := json.NewEncoder(w).Encode(msg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) Run(addr string) {
	http.HandleFunc("/enqueue", s.EnqueueHandler)
	http.HandleFunc("/dequeue", s.DequeueHandler)

	log.Fatal(http.ListenAndServe(addr, nil))
}
