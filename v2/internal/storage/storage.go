package storage

import "context"

// Transaction is a generic interface for a transaction.
type Transaction interface {
	// Get returns the value for a given key.
	Get(key string) ([]byte, error)
	// Put sets the value for a given key.
	Put(key string, value []byte) error
	// Delete deletes the value for a given key.
	Delete(key string) error
	// Scan scans the database for all keys that match the prefix.
	Scan(prefix string) ([]KV, error)
	// ScanKeys scans the database for all keys that match the prefix.
	ScanKeys(prefix string) ([]string, error)
	// Commit commits the current transaction.
	Commit() error
	// Rollback rolls back the current transaction.
	Rollback()
}

// Storage is a generic interface for a key-value store.
type Storage interface {
	// BeginTx starts a new transaction.
	BeginTx() Transaction
	// Get returns the value for a given key.
	Get(key string) ([]byte, error)
	// Stream streams the database for all keys that match the prefix.
	Stream(ctx context.Context, prefix string, yield func(key string, value []byte) error) error
	// Put sets the value for a given key.
	Put(key string, value []byte) error
	// Delete deletes the value for a given key.
	Delete(key string) error
	// Scan scans the database for all keys that match the prefix.
	Scan(prefix string) ([]KV, error)
	// ScanKeys scans the database for all keys that match the prefix.
	ScanKeys(prefix string) ([]string, error)
	// PrintAllKeys prints all keys in the database.
	PrintAllKeys() error
	// Close closes the storage engine.
	Close() error
}

// NewStorage creates a new storage engine.
func NewStorage(path string) (Storage, error) {
	return newBadgerEngine(path)
}
