package index

// CollectionMetadata stores metadata for a collection.
type CollectionMetadata struct {
	Name          string  `json:"name"`
	Indexes       []Model `json:"indexes"`
	DocumentCount int64   `json:"document_count"`
}
