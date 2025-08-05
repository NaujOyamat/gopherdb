package storage

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/ristretto/v2/z"
	. "github.com/NaujOyamat/gopherdb/internal/consts"
)

// badgerEngine is an implementation of the StorageEngine interface that uses BadgerDB as the underlying storage engine.
type badgerEngine struct {
	db *badger.DB
}

// newBadgerEngine creates a new BadgerDB-based storage engine at the given path.
func newBadgerEngine(path string) (*badgerEngine, error) {
	if err := os.MkdirAll(path, P0755); err != nil {
		return nil, err
	}

	opts := badger.DefaultOptions(path).
		WithValueLogFileSize(128 << 20)
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	if db.IsClosed() {
		return nil, ErrDatabaseClosed
	}

	return &badgerEngine{db: db}, nil
}

// BeginTx starts a new transaction.
func (e *badgerEngine) BeginTx() Transaction {
	return newBadgerTransaction(e.db.NewTransaction(true))
}

// Put inserts a key-value pair into the storage engine.
func (e *badgerEngine) Put(key string, value []byte) error {
	if value == nil {
		value = []byte{}
	}

	return e.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
}

// Stream streams the database for all keys that match the prefix.
func (e *badgerEngine) Stream(ctx context.Context, prefix string, yield func(key string, value []byte) error) error {
	stream := e.db.NewStream()
	stream.Prefix = []byte(prefix)
	stream.Send = func(buf *z.Buffer) error {
		if !buf.IsEmpty() {
			kvList, err := badger.BufferToKVList(buf)
			if err != nil {
				return err
			}

			for _, kv := range kvList.Kv {
				if err := yield(string(kv.Key), kv.Value); err != nil {
					return err
				}
			}

			return nil
		}

		return nil
	}

	if err := stream.Orchestrate(ctx); err != nil {
		return fmt.Errorf("stream execution failed: %w", err)
	}

	return nil
}

// Get retrieves a value from the storage engine for the given key.
func (e *badgerEngine) Get(key string) ([]byte, error) {
	var value []byte

	err := e.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrKeyNotFound
			}

			return err
		}

		value, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return value, nil
}

// Delete removes a key-value pair from the storage engine.
func (e *badgerEngine) Delete(key string) error {
	return e.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// Scan scans the storage engine for all keys that match the given prefix.
func (e *badgerEngine) Scan(prefix string) ([]KV, error) {
	results := make([]KV, 0)

	opts := badger.DefaultIteratorOptions

	err := e.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				copied := make([]byte, len(v))
				copy(copied, v)
				results = append(results, KV{Key: string(k), Value: copied})

				return nil
			})

			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// PrintAllKeys prints all keys in the storage engine.
func (e *badgerEngine) PrintAllKeys() error {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false

	err := e.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			fmt.Println("Key:", string(k))
		}

		return nil
	})

	return err
}

// ScanKeys scans the storage engine for all keys that match the given prefix.
func (e *badgerEngine) ScanKeys(prefix string) ([]string, error) {
	results := make([]string, 0)

	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false

	err := e.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			k := item.Key()
			results = append(results, string(k))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// Close closes the storage engine.
func (e *badgerEngine) Close() error {
	return e.db.Close()
}
