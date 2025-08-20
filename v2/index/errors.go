package index

import "errors"

var (
	// ErrMissingFieldForIndex is returned when a field is missing for an index.
	ErrMissingFieldForIndex = errors.New("missing field for index")
	// ErrEmptyIndexFields is returned when an index has no fields.
	ErrEmptyIndexFields = errors.New("empty index fields")
	// ErrDuplicateIndexField is returned when a duplicate index field is found.
	ErrDuplicateIndexField = errors.New("duplicate index field")
	// ErrUniqueIndexViolation is returned when a unique index is violated.
	ErrUniqueIndexViolation = errors.New("unique index violation")
	// ErrIndexAlreadyExists is returned when an index already exists.
	ErrIndexAlreadyExists = errors.New("index already exists")
)
