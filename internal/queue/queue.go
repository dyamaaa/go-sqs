package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
)

// ErrDBOperation is a custom error for database operation failures
var ErrDBOperation = errors.New("database operation failed")

// Message represents a message in the queue
type Message struct {
	ID      string
	Payload string
	visible time.Time // when the message will be visible
}

// Queue represents a queue of messages
type Queue struct {
	db   *badger.DB
	lock sync.Mutex
}

// NewQueue creates a new queue with a database at the given path
func NewQueue(dbPath string) (*Queue, error) {
	opts := badger.DefaultOptions(dbPath).WithLoggingLevel(badger.WARNING)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDBOperation, err)
	}

	return &Queue{
		db: db,
	}, nil
}

// Enqueue adds a new message to the queue
func (q *Queue) Enqueue(message Message) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	err := q.db.Update(func(txn *badger.Txn) error {
		message.visible = time.Now() // make the message visible
		msgJson, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrDBOperation, err)
		}
		err = txn.Set([]byte(message.ID), msgJson)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrDBOperation, err)
		}
		return nil
	})

	if err != nil {
		log.Printf("enqueue operation failed: %v", err)
		return fmt.Errorf("enqueue operation failed: %w", err)
	}

	return nil
}

// Dequeue removes and returns the first visible message from the queue
func (q *Queue) Dequeue(timeout time.Duration) (*Message, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	var msg *Message
	err := q.db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				err := json.Unmarshal(v, &msg)
				if err != nil {
					return fmt.Errorf("%w: %v", ErrDBOperation, err)
				}
				if msg.visible.Before(time.Now()) {
					msg.visible = time.Now().Add(timeout)
					updatedMsg, _ := json.Marshal(msg)
					txn.Set(item.Key(), updatedMsg)
					return nil
				}
				return nil
			})
			if err != nil {
				return err
			}
			// delete the message
			if err := txn.Delete(item.Key()); err != nil {
				return fmt.Errorf("%w: %v", ErrDBOperation, err)
			}
			return nil
		}
		return badger.ErrKeyNotFound
	})

	return msg, err
}

// WaitForMessage waits for a message to become available in the queue
func (q *Queue) WaitForMessage(ctx context.Context) (*Message, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			q.lock.Lock()
			msg, err := q.Dequeue(10 * time.Second)
			if err != nil {
				q.lock.Unlock()
				return nil, err
			}
			if msg != nil {
				q.lock.Unlock()
				return msg, nil
			}
			q.lock.Unlock()
			time.Sleep(100 * time.Millisecond)
		}
	}
}
