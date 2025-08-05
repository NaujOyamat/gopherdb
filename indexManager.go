package gopherdb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/NaujOyamat/gopherdb/internal/bson"
	"github.com/NaujOyamat/gopherdb/internal/consts"
	"github.com/NaujOyamat/gopherdb/internal/storage"
)

type indexTask struct {
	doc map[string]any
}

type worker struct {
	id     int
	parent *IndexManager
}

// IndexManager is a manager for indexes.
type IndexManager struct {
	dbname   string
	collname string
	storage  storage.Storage
	metadata CollectionMetadata
	mu       sync.Mutex
}

// newIndexManager creates a new IndexManager.
func newIndexManager(storage storage.Storage, dbname, collname string) *IndexManager {
	return &IndexManager{
		mu:       sync.Mutex{},
		storage:  storage,
		dbname:   dbname,
		collname: collname,
	}
}

// List returns all the indexes for the collection.
func (m *IndexManager) List() []IndexModel {
	m.loadMetadata()

	return m.metadata.Indexes
}

// buildMetadataKey builds the key for the collection metadata.
func (m *IndexManager) buildMetadataKey() string {
	return fmt.Sprintf(consts.MetadataCollectionKeyStringFormat, m.dbname, m.collname)
}

// loadMetadata loads the collection metadata from the storage.
func (m *IndexManager) loadMetadata() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := m.storage.Get(m.buildMetadataKey())
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			m.metadata = CollectionMetadata{
				Name: fmt.Sprintf(
					consts.CollectionKeyStringFormat,
					m.dbname,
					m.collname,
				),
				Indexes: []IndexModel{
					{
						Fields: []IndexField{
							{
								Name:  consts.DocumentFieldID,
								Order: 1,
							},
						},
						Options: IndexOptions{
							Name:   "_id_",
							Unique: true,
						},
					},
				},
				DocumentCount: 0,
			}

			return nil
		}

		return err
	}

	return bson.Unmarshal(data, &m.metadata)
}

// saveMetadata saves the collection metadata to the storage.
func (m *IndexManager) saveMetadata() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := bson.Marshal(m.metadata)
	if err != nil {
		return err
	}

	return m.storage.Put(m.buildMetadataKey(), data)
}

// getDocumentIdFromIndexKey gets the document id from an index key.
func (m *IndexManager) getDocumentIdFromIndexKey(indexKey string) (string, error) {
	match, err := consts.IndexKeyPathmatcher.Match(indexKey)
	if err != nil {
		return "", err
	}

	return match["docId"], nil
}

// getDocumentIndexKeysByIndex gets all the document ids for a given index.
func (m *IndexManager) getDocumentIndexKeysByIndex(index IndexModel) ([]string, error) {
	indexKeyPrefix := m.buildIndexFieldsKey(index)

	entries, err := m.storage.ScanKeys(indexKeyPrefix)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// getDocumentIndexKeysByIndexAndFilter gets the document ids for a given index.
func (m *IndexManager) getDocumentIndexKeysByIndexAndFilter(index IndexModel, indexFilter map[string]any) ([]string, error) {
	indexKeyPrefix, err := m.buildDocumentIndexKey(index, indexFilter, true)
	if err != nil {
		return nil, err
	}

	indexKeyPrefix = strings.TrimSuffix(
		indexKeyPrefix,
		fmt.Sprintf("/%v", consts.RemoverWildcard),
	)

	entries, err := m.storage.ScanKeys(indexKeyPrefix)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// buildIndexFieldsKey builds the fields key for a given index.
func (m *IndexManager) buildIndexFieldsKey(index IndexModel) string {
	fields := make([]string, 0, len(index.Fields))

	for _, f := range index.Fields {
		fields = append(fields, f.Name)
	}

	joinedFields := strings.Join(fields, "|")

	key := fmt.Sprintf(
		consts.IndexKeyStringFormat,
		m.dbname,
		m.collname,
		index.Options.Name,
		joinedFields,
		consts.RemoverWildcard,
		consts.RemoverWildcard,
	)

	return strings.TrimSuffix(
		key,
		fmt.Sprintf(
			"%v/%v",
			consts.RemoverWildcard,
			consts.RemoverWildcard,
		),
	)
}

// buildDocumentIndexKey builds the index key for a document.
func (m *IndexManager) buildDocumentIndexKey(index IndexModel, doc map[string]any, isPrefix bool) (string, error) {
	values := make([]string, 0, len(index.Fields))
	fields := make([]string, 0, len(index.Fields))

	for _, f := range index.Fields {
		fields = append(fields, f.Name)

		val, ok := doc[f.Name]
		if ok {
			values = append(values, encodeForLexOrder(val, f.Order < 0))
		} else {
			if isPrefix {
				continue
			}

			return "", ErrMissingFieldForIndex
		}
	}

	joinedFields := strings.Join(fields, "|")
	joinedValues := strings.Join(values, "|")

	docId := fmt.Sprintf("%v", doc[consts.DocumentFieldID])
	if isPrefix {
		docId = consts.RemoverWildcard
	}

	return fmt.Sprintf(
		consts.IndexKeyStringFormat,
		m.dbname,
		m.collname,
		index.Options.Name,
		joinedFields,
		joinedValues,
		docId,
	), nil
}

// checkUniqueness checks if the document violates the uniqueness constraint of the index.
func (m *IndexManager) checkUniqueness(doc map[string]any) error {
	for _, idx := range m.metadata.Indexes {
		if !idx.isUnique() {
			continue
		}

		idxKey, err := m.buildDocumentIndexKey(idx, doc, false)
		if err != nil {
			return err
		}

		key := strings.TrimSuffix(idxKey, fmt.Sprintf("%v", doc[consts.DocumentFieldID]))
		entries, err := m.storage.ScanKeys(key)

		if err != nil {
			return err
		}

		if len(entries) > 0 {
			return fmt.Errorf("%w: fields %+v", ErrUniqueIndexViolation, idx.Fields)
		}
	}

	return nil
}

// CreateMany creates many indexes.
func (m *IndexManager) CreateMany(ctx context.Context, indexes []IndexModel) error {
	if indexes == nil || len(indexes) == 0 {
		return nil
	}

	m.loadMetadata()
	defer m.buildIndexes(ctx)

	indexes = splitCompoundIndexes(indexes)

	for _, newidx := range indexes {
		if err := newidx.validate(); err != nil {
			return err
		}

		for _, idx := range m.metadata.Indexes {
			if idx.Options.Name == newidx.Options.Name {
				if newidx.isAutogenerated() {
					continue
				}

				if !idx.isAutogenerated() {
					return fmt.Errorf("%w: index name %s", ErrIndexAlreadyExists, newidx.Options.Name)
				}
			}

			found := 0

			for _, newf := range newidx.Fields {
				for _, f := range idx.Fields {
					if f.Name == newf.Name {
						found++
					}
				}
			}

			if found == len(newidx.Fields) {
				if newidx.isAutogenerated() {
					continue
				}

				if !idx.isAutogenerated() {
					return fmt.Errorf("%w: index fields %v", ErrIndexAlreadyExists, newidx.Fields)
				}

				idx.Options = newidx.Options
				idx.Options.Autogenerated = false
				idx.Fields = newidx.Fields

				continue
			}
		}

		m.metadata.Indexes = append(m.metadata.Indexes, newidx)
	}

	return m.saveMetadata()
}

// indexDocument indexes a document.
func (m *IndexManager) indexDocument(txn storage.Transaction, doc map[string]any) error {
	for _, idx := range m.metadata.Indexes {
		idxKey, err := m.buildDocumentIndexKey(idx, doc, false)
		if err != nil {
			return err
		}

		if err := txn.Put(idxKey, nil); err != nil {
			return err
		}
	}

	return nil
}

// deleteDocumentIndexes deletes the indexes for a document.
func (m *IndexManager) deleteDocumentIndexes(doc map[string]any) error {
	for _, idx := range m.metadata.Indexes {
		idxKey, err := m.buildDocumentIndexKey(idx, doc, false)
		if err != nil {
			return err
		}

		if err := m.storage.Delete(idxKey); err != nil {
			return err
		}
	}

	return nil
}

// buildDocumentKey builds the document key.
func (m *IndexManager) buildDocumentKey(docID string) string {
	return fmt.Sprintf(consts.DocumentKeyStringFormat, m.dbname, m.collname, docID)
}

// buildDocumentsKey builds the documents key.
func (m *IndexManager) buildDocumentsKey() string {
	key := fmt.Sprintf(consts.DocumentKeyStringFormat, m.dbname, m.collname, consts.RemoverWildcard)

	return strings.TrimSuffix(
		key,
		consts.RemoverWildcard,
	)
}

// buildIndexes builds the indexes for a collection.
func (m *IndexManager) buildIndexes(ctx context.Context) {
	go func() {
		m.loadMetadata()

		docsPrefix := strings.TrimSuffix(
			m.buildDocumentKey(consts.RemoverWildcard),
			consts.RemoverWildcard,
		)

		docs, err := m.storage.Scan(docsPrefix)
		if err != nil {
			return
		}

		txn := m.storage.BeginTx()

		for _, doc := range docs {
			select {
			case <-ctx.Done():
				txn.Rollback()

				return
			default:
				var docMap map[string]any
				if err := bson.Unmarshal(doc.Value, &docMap); err != nil {
					return
				}

				if err := m.indexDocument(txn, docMap); err != nil {
					return
				}
			}
		}

		if err := txn.Commit(); err != nil {
			return
		}
	}()
}
