package queue

import (
	"encoding/json"
	"sync"

	"github.com/dgraph-io/badger/v3"
)

type Message struct {
	ID      string
	Payload string
}

type Queue struct {
	db   *badger.DB
	lock sync.Mutex
}

func NewQueue(dbPath string) (*Queue, error) {
	opts := badger.DefaultOptions(dbPath).WithLoggingLevel(badger.WARNING)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
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
			return err
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
				return err
			})
			if err != nil {
				return err
			}
			// メッセージを読み込んだ後に削除
			if err := txn.Delete(item.Key()); err != nil {
				return err
			}
			return nil
		}
		return badger.ErrKeyNotFound
	})

	return msg, err
}
