package storage

import "github.com/NaujOyamat/gopherdb/internal/bson"

// KV is a key-value pair.
type KV struct {
	Key   string
	Value []byte
}

// Document returns the document of the KV.
func (k *KV) Document() map[string]any {
	var doc map[string]any
	err := bson.Unmarshal(k.Value, &doc)

	if err != nil {
		return nil
	}

	return doc
}
