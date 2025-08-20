package gopherdb

import indexpkg "github.com/NaujOyamat/gopherdb/v2/index"

type (
	// IndexOptions represents the options for an index.
	IndexOptions = indexpkg.Options
	// IndexField represents a field in an index.
	IndexField = indexpkg.Field
	// IndexModel represents a model for an index.
	IndexModel = indexpkg.Model
	// IndexManager manages collection indexes.
	IndexManager = indexpkg.Manager
)

// NewIndexOptions creates a new index options.
func NewIndexOptions() *IndexOptions { return indexpkg.NewOptions() }

// NewIndexModel creates a new index model.
func NewIndexModel() *IndexModel { return indexpkg.NewModel() }
