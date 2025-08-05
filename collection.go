package gopherdb

import (
	"cmp"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/NaujOyamat/gopherdb/internal/bson"
	"github.com/NaujOyamat/gopherdb/internal/consts"
	"github.com/NaujOyamat/gopherdb/internal/storage"
	"github.com/NaujOyamat/gopherdb/options"
)

// Collection is a collection of documents.
type Collection struct {
	dbname       string
	collname     string
	storage      storage.Storage
	initialized  bool
	IndexManager *IndexManager
}

// newCollection creates a new collection.
func newCollection(storage storage.Storage, dbname, collname string) (*Collection, error) {
	idxMgr := newIndexManager(storage, dbname, collname)
	if err := idxMgr.loadMetadata(); err != nil {
		return nil, err
	}

	return &Collection{
		dbname:       dbname,
		collname:     collname,
		storage:      storage,
		initialized:  (idxMgr.metadata.DocumentCount > 0),
		IndexManager: idxMgr,
	}, nil
}

// ensureInitialized ensures that the collection is initialized.
func (c *Collection) ensureInitialized() error {
	if c.initialized {
		return nil
	}

	if c.IndexManager.metadata.DocumentCount > 0 {
		if err := c.IndexManager.saveMetadata(); err != nil {
			return err
		}

		c.initialized = true
	}

	return nil
}

// buildDocumentKey builds the key for a document.
func (c *Collection) buildDocumentKey(docID string) string {
	return fmt.Sprintf(consts.DocumentKeyStringFormat, c.dbname, c.collname, docID)
}

// ensureDocumentID ensures that the document has an ID.
func (c *Collection) ensureDocumentID(doc map[string]any) (map[string]any, string) {
	copyDoc := make(map[string]any)
	maps.Copy(copyDoc, doc)
	val, ok := copyDoc[consts.DocumentFieldID]

	if !ok {
		copyDoc[consts.DocumentFieldID] = uuid.NewString()
	} else {
		tval := reflect.ValueOf(val)
		if tval.IsZero() && tval.Kind() == reflect.String {
			copyDoc[consts.DocumentFieldID] = uuid.NewString()
		}
	}

	return copyDoc, fmt.Sprintf("%v", copyDoc[consts.DocumentFieldID])
}

// updateOne updates a single document by a filter.
func (c *Collection) updateOne(
	txn storage.Transaction,
	filter map[string]any,
	doc any,
	opts ...*options.UpdateOptions,
) UpdateOneResult {
	c.IndexManager.loadMetadata()

	if doc == nil {
		return UpdateOneResult{
			Err: ErrDocumentIsNil,
		}
	}

	docVal := reflect.ValueOf(doc)
	if docVal.Kind() == reflect.Ptr {
		if docVal.IsNil() {
			return UpdateOneResult{
				Err: ErrDocumentPointerIsNil,
			}
		}

		docVal = docVal.Elem()
	}

	docKind := docVal.Kind()
	if docKind != reflect.Map && docKind != reflect.Struct {
		return UpdateOneResult{
			Err: ErrDocumentTypeInvalid,
		}
	}

	if docKind == reflect.Map {
		mapType := docVal.Type()
		if mapType.Key().Kind() != reflect.String || mapType.Elem().Kind() != reflect.Interface {
			return UpdateOneResult{
				Err: ErrDocumentTypeInvalid,
			}
		}
	}

	opt := options.Update()
	if len(opts) > 0 {
		opt = opt.Merge(opts...)
	}

	result := c.FindOne(filter)

	if result.Err != nil {
		if result.Err == ErrDocumentNotFound && opt.Upsert != nil && *opt.Upsert {
			insertResult := c.insertOne(txn, doc)
			if insertResult.Err != nil {
				return UpdateOneResult{
					Err: insertResult.Err,
				}
			}

			return UpdateOneResult{
				UpsertedID: insertResult.InsertedID,
			}
		}

		return UpdateOneResult{
			Err: result.Err,
		}
	}

	match, err := consts.DocumentKeyPathmatcher.Match(result.raw.Key)
	if err != nil {
		return UpdateOneResult{
			Err: fmt.Errorf("match failed: %w", err),
		}
	}

	docID := match["docId"]

	docMap, err := bson.ConvertToMap(doc)
	if err != nil {
		return UpdateOneResult{
			Err: fmt.Errorf("error converting document to map: %w", err),
		}
	}

	docMapID, docMapIDOk := docMap[consts.DocumentFieldID]
	if !docMapIDOk {
		docMap[consts.DocumentFieldID] = docID
		docMapID = docID
	}

	if docID != docMapID {
		return UpdateOneResult{
			Err: ErrDocumentIDNoEditable,
		}
	}

	docUpdate := make(map[string]any)
	if opt.Set != nil && *opt.Set {
		maps.Copy(docUpdate, docMap)
	} else {
		maps.Copy(docUpdate, result.Document())
		maps.Copy(docUpdate, docMap)
	}

	bdoc, err := bson.Marshal(docUpdate)
	if err != nil {
		return UpdateOneResult{
			Err: fmt.Errorf("bson marshal failed: %w", err),
		}
	}

	if err := txn.Put(result.raw.Key, bdoc); err != nil {
		return UpdateOneResult{
			Err: fmt.Errorf("update failed: %w", err),
		}
	}

	err = c.IndexManager.indexDocument(txn, docUpdate)
	if err != nil {
		return UpdateOneResult{
			Err: fmt.Errorf("index document failed: %w", err),
		}
	}

	return UpdateOneResult{
		UpsertedID: docID,
	}
}

// insertOne inserts a single document into the collection.
func (c *Collection) insertOne(txn storage.Transaction, doc any) InsertOneResult {
	c.IndexManager.loadMetadata()

	// 1. Convertimos a BSON (map[string]interface{})
	parsed, err := bson.ConvertToMap(doc)
	if err != nil {
		return InsertOneResult{
			Err: fmt.Errorf("bson conversion failed: %w", err),
		}
	}

	// 2. Generamos ID único
	mDoc, docID := c.ensureDocumentID(parsed)

	// 3. Verificamos unicidad en índices
	if err := c.IndexManager.checkUniqueness(mDoc); err != nil {
		return InsertOneResult{
			Err: err,
		}
	}

	// 4. Serializamos a JSON
	data, err := bson.Marshal(mDoc)
	if err != nil {
		return InsertOneResult{
			Err: fmt.Errorf("json marshal failed: %w", err),
		}
	}

	if err := bson.ValidateBSON(data); err != nil {
		return InsertOneResult{
			Err: fmt.Errorf("bson validation failed: %w", err),
		}
	}

	// 5. Guardamos el documento
	key := c.buildDocumentKey(docID)
	if err := txn.Put(key, data); err != nil {
		return InsertOneResult{
			Err: fmt.Errorf("storage put failed: %w", err),
		}
	}

	c.IndexManager.metadata.DocumentCount++

	// 6. Registramos índices secundarios
	err = c.IndexManager.indexDocument(txn, mDoc)
	if err != nil {
		return InsertOneResult{
			Err: fmt.Errorf("index document failed: %w", err),
		}
	}

	// 7. Persistimos metadata si es la primera vez
	if err := c.ensureInitialized(); err != nil {
		return InsertOneResult{
			Err: fmt.Errorf("metadata initialization failed: %w", err),
		}
	}

	return InsertOneResult{
		InsertedID: docID,
	}
}

// deleteOne deletes a single document by a filter.
func (c *Collection) deleteOne(txn storage.Transaction, filter map[string]any) DeleteOneResult {
	c.IndexManager.loadMetadata()

	result := c.FindOne(filter)
	if result.Err != nil {
		return DeleteOneResult{
			Err: result.Err,
		}
	}

	match, err := consts.DocumentKeyPathmatcher.Match(result.raw.Key)
	if err != nil {
		return DeleteOneResult{
			Err: fmt.Errorf("match failed: %w", err),
		}
	}

	docID := match["docId"]
	key := c.buildDocumentKey(docID)

	if err := txn.Delete(key); err != nil {
		return DeleteOneResult{
			Err: fmt.Errorf("delete failed: %w", err),
		}
	}

	c.IndexManager.metadata.DocumentCount--

	err = c.IndexManager.deleteDocumentIndexes(result.Document())
	if err != nil {
		return DeleteOneResult{
			Err: fmt.Errorf("delete document indexes failed: %w", err),
		}
	}

	c.IndexManager.saveMetadata()

	return DeleteOneResult{
		DeletedID: docID,
	}
}

// sortDocuments sorts the documents by the given sort options.
func (c *Collection) sortDocuments(docs []storage.KV, opt *options.FindOptions) {
	slices.SortStableFunc(docs, func(a, b storage.KV) int {
		for _, f := range opt.Sort {
			va, aok := a.Document()[f.Field]
			vb, bok := b.Document()[f.Field]

			if va == nil || vb == nil || !aok || !bok {
				continue
			}

			var result int

			switch vaTyped := va.(type) {
			case string:
				if vbTyped, ok := vb.(string); ok {
					result = cmp.Compare(strings.ToLower(vaTyped), strings.ToLower(vbTyped))
				}
			case int:
				if vbTyped, ok := vb.(int); ok {
					result = cmp.Compare(vaTyped, vbTyped)
				}
			case int8:
				if vbTyped, ok := vb.(int8); ok {
					result = cmp.Compare(vaTyped, vbTyped)
				}
			case int16:
				if vbTyped, ok := vb.(int16); ok {
					result = cmp.Compare(vaTyped, vbTyped)
				}
			case int32:
				if vbTyped, ok := vb.(int32); ok {
					result = cmp.Compare(vaTyped, vbTyped)
				}
			case int64:
				if vbTyped, ok := vb.(int64); ok {
					result = cmp.Compare(vaTyped, vbTyped)
				}
			case float32:
				if vbTyped, ok := vb.(float32); ok {
					result = cmp.Compare(vaTyped, vbTyped)
				}
			case float64:
				if vbTyped, ok := vb.(float64); ok {
					result = cmp.Compare(vaTyped, vbTyped)
				}
			case bool:
				if vbTyped, ok := vb.(bool); ok {
					vaInt, vbInt := 0, 0
					if vaTyped {
						vaInt = 1
					}

					if vbTyped {
						vbInt = 1
					}

					result = cmp.Compare(vaInt, vbInt)
				}
			case time.Time:
				if vbTyped, ok := vb.(time.Time); ok {
					result = cmp.Compare(vaTyped.UnixNano(), vbTyped.UnixNano())
				}
			case time.Duration:
				if vbTyped, ok := vb.(time.Duration); ok {
					result = cmp.Compare(vaTyped, vbTyped)
				}
			case uint8:
				if vbTyped, ok := vb.(uint8); ok {
					result = cmp.Compare(vaTyped, vbTyped)
				}
			case uint16:
				if vbTyped, ok := vb.(uint16); ok {
					result = cmp.Compare(vaTyped, vbTyped)
				}
			case uint32:
				if vbTyped, ok := vb.(uint32); ok {
					result = cmp.Compare(vaTyped, vbTyped)
				}
			case uint64:
				if vbTyped, ok := vb.(uint64); ok {
					result = cmp.Compare(vaTyped, vbTyped)
				}
			case uintptr:
				if vbTyped, ok := vb.(uintptr); ok {
					result = cmp.Compare(vaTyped, vbTyped)
				}
			default:
				continue
			}

			if result != 0 {
				if f.Order < 0 {
					return -result
				}

				return result
			}
		}

		return 0
	})
}
