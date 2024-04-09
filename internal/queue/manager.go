package queue

import (
	"fmt"
	"sync"
)

type Manager struct {
	queues map[string]*Queue
	lock   sync.RWMutex
	dbPath string
}

func NewManager(dbPath string) (*Manager, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("dbPath cannot be empty")
	}

	return &Manager{
		queues: make(map[string]*Queue),
		dbPath: dbPath,
	}, nil
}

func (m *Manager) GetOrCreateQueue(name string) (*Queue, error) {
	queue, exists := m.getQueue(name)
	if exists {
		return queue, nil
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	queue, exists = m.getQueue(name)
	if exists {
		return queue, nil
	}

	queuePath := fmt.Sprintf("%s/%s", m.dbPath, name)
	queue, err := NewQueue(queuePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create new queue: %w", err)
	}

	m.queues[name] = queue
	return queue, nil
}

// getQueue is a helper method to get a queue from the map if it exists.
func (m *Manager) getQueue(name string) (*Queue, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	queue, exists := m.queues[name]
	return queue, exists
}
