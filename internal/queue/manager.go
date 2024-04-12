package queue

import (
	"fmt"
	"sync"
)

// Manager is a struct that manages multiple queues.
type Manager struct {
	queues map[string]*Queue // queues is a map of queue names to Queue instances.
	lock   sync.RWMutex      // lock is used to ensure thread safety when accessing the queues map.
	dbPath string            // dbPath is the path to the database where the queues are stored.
}

// NewManager creates a new Manager instance.
// It requires a dbPath which cannot be empty.
func NewManager(dbPath string) (*Manager, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("dbPath cannot be empty")
	}

	return &Manager{
		queues: make(map[string]*Queue), // Initialize an empty map for queues.
		dbPath: dbPath,                  // Set the database path.
	}, nil
}

// GetOrCreateQueue gets a queue with the given name if it exists, or creates a new one if it doesn't.
func (m *Manager) GetOrCreateQueue(name string) (*Queue, error) {
	queue, exists := m.getQueue(name)
	if exists {
		return queue, nil
	}

	m.lock.Lock() // Lock the queues map for writing.
	defer m.lock.Unlock()

	queue, exists = m.getQueue(name)
	if exists {
		return queue, nil
	}

	queuePath := fmt.Sprintf("%s/%s", m.dbPath, name)
	queue, err := NewQueue(queuePath) // Create a new queue.
	if err != nil {
		return nil, fmt.Errorf("failed to create new queue: %w", err)
	}

	m.queues[name] = queue // Add the new queue to the queues map.
	return queue, nil
}

// getQueue is a helper method to get a queue from the map if it exists.
// It locks the queues map for reading.
func (m *Manager) getQueue(name string) (*Queue, bool) {
	m.lock.RLock() // Lock the queues map for reading.
	defer m.lock.RUnlock()
	queue, exists := m.queues[name]
	return queue, exists
}
