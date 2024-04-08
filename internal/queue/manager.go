package queue

import (
	"fmt"
	"sync"
)

type Manager struct {
	queues map[string]*Queue
	lock   sync.RWMutex
	dbPath string // データベースファイルのベースパス
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
	m.lock.RLock()
	queue, exists := m.queues[name]
	m.lock.RUnlock()

	if exists {
		return queue, nil
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	// Check again in case another goroutine created the queue while we were waiting for the lock
	if queue, exists := m.queues[name]; exists {
		return queue, nil
	}

	// キューごとに一意のデータベースパスを生成
	queuePath := fmt.Sprintf("%s/%s", m.dbPath, name)
	queue, err := NewQueue(queuePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create new queue: %w", err)
	}

	m.queues[name] = queue
	return queue, nil
}
