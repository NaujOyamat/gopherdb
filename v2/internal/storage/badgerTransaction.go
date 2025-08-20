package storage

import (
	"errors"

	"github.com/dgraph-io/badger/v4"
)

// badgerTransaction is a transaction for the badger storage engine.
type badgerTransaction struct {
	txn *badger.Txn
}

// newBadgerTransaction creates a new badger transaction.
func newBadgerTransaction(txn *badger.Txn) *badgerTransaction {
	return &badgerTransaction{txn: txn}
}

// Get returns the value for a given key.
func (t *badgerTransaction) Get(key string) ([]byte, error) {
	item, err := t.txn.Get([]byte(key))
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, ErrKeyNotFound
		}

		return nil, err
	}

	value, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// Put sets the value for a given key.
func (t *badgerTransaction) Put(key string, value []byte) error {
	return t.txn.Set([]byte(key), value)
}

// Delete deletes the value for a given key.
func (t *badgerTransaction) Delete(key string) error {
	return t.txn.Delete([]byte(key))
}

// Scan scans the database for all keys that match the prefix.
func (t *badgerTransaction) Scan(prefix string) ([]KV, error) {
	results := make([]KV, 0)

	opts := badger.DefaultIteratorOptions

	it := t.txn.NewIterator(opts)
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
			return nil, err
		}
	}

	return results, nil
}

// ScanKeys scans the database for all keys that match the prefix.
func (t *badgerTransaction) ScanKeys(prefix string) ([]string, error) {
	results := make([]string, 0)

	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false

	it := t.txn.NewIterator(opts)
	defer it.Close()

	for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
		item := it.Item()
		k := item.Key()
		results = append(results, string(k))
	}

	return results, nil
}

// Commit commits the current transaction.
func (t *badgerTransaction) Commit() error {
	return t.txn.Commit()
}

// Rollback rolls back the current transaction.
func (t *badgerTransaction) Rollback() {
	t.txn.Discard()
}
