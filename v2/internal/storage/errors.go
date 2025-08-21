package storage

import "errors"

var (
	// ErrKeyNotFound is returned when a key is not found in the storage.
	ErrKeyNotFound = errors.New("key not found")
	// ErrDatabaseClosed is returned when a database is closed.
	ErrDatabaseClosed = errors.New("database is closed")
)
