package index

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/NaujOyamat/gopherdb/v2/internal/bson"
	"github.com/NaujOyamat/gopherdb/v2/internal/consts"
	"github.com/NaujOyamat/gopherdb/v2/internal/storage"
)

type indexTask struct {
	doc map[string]any
}

type worker struct {
	id     int
	parent *Manager
}

// Manager is a manager for indexes.
type Manager struct {
	dbname   string
	collname string
	storage  storage.Storage
	Metadata CollectionMetadata
	mu       sync.RWMutex
	loaded   bool
}

// NewManager creates a new Manager and loads metadata.
func NewManager(storage storage.Storage, dbname, collname string) (*Manager, error) {
	m := &Manager{
		mu:       sync.RWMutex{},
		storage:  storage,
		dbname:   dbname,
		collname: collname,
	}

	if err := m.LoadMetadata(); err != nil {
		return nil, err
	}

	return m, nil
}

// List returns all the indexes for the collection.
func (m *Manager) List() []Model {
	m.LoadMetadata()

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.Metadata.Indexes
}

// DocumentCount returns the number of documents tracked by the collection.
func (m *Manager) DocumentCount() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.Metadata.DocumentCount
}

// IncrementDocumentCount increases the document count by one.
func (m *Manager) IncrementDocumentCount() {
	m.mu.Lock()
	m.Metadata.DocumentCount++
	m.mu.Unlock()
}

// DecrementDocumentCount decreases the document count by one.
func (m *Manager) DecrementDocumentCount() {
	m.mu.Lock()
	m.Metadata.DocumentCount--
	m.mu.Unlock()
}

// buildMetadataKey builds the key for the collection metadata.
func (m *Manager) buildMetadataKey() string {
	return fmt.Sprintf(consts.MetadataCollectionKeyStringFormat, m.dbname, m.collname)
}

// LoadMetadata loads the collection metadata from the storage.
func (m *Manager) LoadMetadata() error {
	m.mu.RLock()
	if m.loaded {
		m.mu.RUnlock()
		return nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.loaded {
		return nil
	}

	data, err := m.storage.Get(m.buildMetadataKey())
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			m.Metadata = CollectionMetadata{
				Name: fmt.Sprintf(
					consts.CollectionKeyStringFormat,
					m.dbname,
					m.collname,
				),
				Indexes: []Model{
					{
						Fields: []Field{
							{
								Name:  consts.DocumentFieldID,
								Order: 1,
							},
						},
						Options: Options{
							Name:   "_id_",
							Unique: true,
						},
					},
				},
				DocumentCount: 0,
			}
			m.loaded = true

			return nil
		}

		return err
	}

	if err := bson.Unmarshal(data, &m.Metadata); err != nil {
		return err
	}

	m.loaded = true

	return nil
}

// SaveMetadata saves the collection metadata to the storage.
func (m *Manager) saveMetadataLocked() error {
	data, err := bson.Marshal(m.Metadata)
	if err != nil {
		return err
	}

	return m.storage.Put(m.buildMetadataKey(), data)
}

// SaveMetadata saves the collection metadata to the storage.
func (m *Manager) SaveMetadata() error {
	if err := m.LoadMetadata(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	return m.saveMetadataLocked()
}

// GetDocumentIDFromIndexKey gets the document id from an index key.
func (m *Manager) GetDocumentIDFromIndexKey(indexKey string) (string, error) {
	match, err := consts.IndexKeyPathmatcher.Match(indexKey)
	if err != nil {
		return "", err
	}

	return match["docId"], nil
}

// GetDocumentIndexKeysByIndex gets all the document ids for a given index.
func (m *Manager) GetDocumentIndexKeysByIndex(index Model) ([]string, error) {
	indexKeyPrefix := m.buildIndexFieldsKey(index)

	entries, err := m.storage.ScanKeys(indexKeyPrefix)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// GetDocumentIndexKeysByIndexAndFilter gets the document ids for a given index.
func (m *Manager) GetDocumentIndexKeysByIndexAndFilter(index Model, indexFilter map[string]any) ([]string, error) {
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
func (m *Manager) buildIndexFieldsKey(index Model) string {
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
func (m *Manager) buildDocumentIndexKey(index Model, doc map[string]any, isPrefix bool) (string, error) {
	values := make([]string, 0, len(index.Fields))
	fields := make([]string, 0, len(index.Fields))

	for _, f := range index.Fields {
		fields = append(fields, f.Name)

               val, ok := doc[f.Name]
               if ok {
                       enc, err := encodeForLexOrder(val, f.Order < 0)
                       if err != nil {
                               return "", err
                       }
                       values = append(values, enc)
               } else {
                       if isPrefix {
                               continue
                       }

			return "", ErrMissingFieldForIndex
		}
	}

	joinedFields := strings.Join(fields, "|")
	joinedValues := strings.Join(values, "|")

	docID := fmt.Sprintf("%v", doc[consts.DocumentFieldID])
	if isPrefix {
		docID = consts.RemoverWildcard
	}

	return fmt.Sprintf(
		consts.IndexKeyStringFormat,
		m.dbname,
		m.collname,
		index.Options.Name,
		joinedFields,
		joinedValues,
		docID,
	), nil
}

// CheckUniqueness checks if the document violates the uniqueness constraint of the index.
func (m *Manager) CheckUniqueness(doc map[string]any) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, idx := range m.Metadata.Indexes {
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
func (m *Manager) CreateMany(ctx context.Context, indexes []Model) error {
	if indexes == nil || len(indexes) == 0 {
		return nil
	}

	if err := m.LoadMetadata(); err != nil {
		return err
	}
	defer m.buildIndexes(ctx)

	indexes = splitCompoundIndexes(indexes)
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, newidx := range indexes {
		if err := newidx.validate(); err != nil {
			return err
		}

		for i := range m.Metadata.Indexes {
			idx := &m.Metadata.Indexes[i]
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

		m.Metadata.Indexes = append(m.Metadata.Indexes, newidx)
	}

	return m.saveMetadataLocked()
}

// IndexDocument indexes a document.
func (m *Manager) IndexDocument(txn storage.Transaction, doc map[string]any) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, idx := range m.Metadata.Indexes {
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

// DeleteDocumentIndexes deletes the indexes for a document.
func (m *Manager) DeleteDocumentIndexes(doc map[string]any) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, idx := range m.Metadata.Indexes {
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
func (m *Manager) buildDocumentKey(docID string) string {
	return fmt.Sprintf(consts.DocumentKeyStringFormat, m.dbname, m.collname, docID)
}

// BuildDocumentsKey builds the documents key.
func (m *Manager) BuildDocumentsKey() string {
	key := fmt.Sprintf(consts.DocumentKeyStringFormat, m.dbname, m.collname, consts.RemoverWildcard)

	return strings.TrimSuffix(
		key,
		consts.RemoverWildcard,
	)
}

// buildIndexes builds the indexes for a collection.
func (m *Manager) buildIndexes(ctx context.Context) {
	go func() {
		if err := m.LoadMetadata(); err != nil {
			return
		}

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

				if err := m.IndexDocument(txn, docMap); err != nil {
					return
				}
			}
		}

		if err := txn.Commit(); err != nil {
			return
		}
	}()
}
