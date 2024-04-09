package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
)

var ErrDBOpen = errors.New("failed to open database")
var ErrMarshal = errors.New("failed to marshal message")
var ErrUnmarshal = errors.New("failed to unmarshal message")
var ErrDelete = errors.New("failed to delete message")

type Message struct {
	ID      string
	Payload string
}

type Queue struct {
	db   *badger.DB
	lock sync.RWMutex
}

func NewQueue(dbPath string) (*Queue, error) {
	opts := badger.DefaultOptions(dbPath).WithLoggingLevel(badger.WARNING)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDBOpen, err)
	}

	return &Queue{
		db: db,
	}, nil
}

func (q *Queue) Enqueue(message Message) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	err := q.db.Update(func(txn *badger.Txn) error {
		msgJson, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrMarshal, err)
		}
		return txn.Set([]byte(message.ID), msgJson)
	})

	return err
}

func (q *Queue) Dequeue() (*Message, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	var msg *Message
	err := q.db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		if it.Rewind(); it.Valid() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				err := json.Unmarshal(v, &msg)
				if err != nil {
					return fmt.Errorf("%w: %v", ErrUnmarshal, err)
				}
				return nil
			})
			if err != nil {
				return err
			}
			// delete the message
			if err := txn.Delete(item.Key()); err != nil {
				return fmt.Errorf("%w: %v", ErrDelete, err)
			}
			return nil
		}
		return badger.ErrKeyNotFound
	})

	return msg, err
}

func (q *Queue) Poll(duration time.Duration, interval time.Duration) (*Message, error) {
	endTime := time.Now().Add(duration)
	for {
		if time.Now().After(endTime) {
			return nil, nil
		}

		msg, err := q.Dequeue()
		if err != nil && err != badger.ErrKeyNotFound {
			return nil, err
		}
		if msg != nil {
			return msg, nil
		}

		time.Sleep(interval)
	}
}

func (q *Queue) WaitForMessage(timeout time.Duration) (*Message, error) {
	start := time.Now()
	for {
		q.lock.Lock()
		msg, err := q.Dequeue()
		if err != nil {
			q.lock.Unlock()
			return nil, err
		}
		if msg != nil {
			q.lock.Unlock()
			return msg, nil
		}
		q.lock.Unlock()

		if time.Since(start) > timeout {
			return nil, errors.New("timeout")
		}

		time.Sleep(100 * time.Millisecond)
	}
}
