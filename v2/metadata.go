package gopherdb

import indexpkg "github.com/NaujOyamat/gopherdb/v2/index"

type (
	// DatabaseMetadata stores metadata for a database.
	DatabaseMetadata struct {
		Name string `json:"name"`
	}
	// CollectionMetadata stores metadata for a collection.
	CollectionMetadata = indexpkg.CollectionMetadata
)
