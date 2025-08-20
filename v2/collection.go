package gopherdb

import (
        "cmp"
        "fmt"
        "reflect"
        "slices"
        "strings"
        "time"

	idx "github.com/NaujOyamat/gopherdb/v2/index"
	"github.com/NaujOyamat/gopherdb/v2/internal/bson"
	"github.com/NaujOyamat/gopherdb/v2/internal/consts"
	"github.com/NaujOyamat/gopherdb/v2/internal/storage"
	"github.com/NaujOyamat/gopherdb/v2/options"
	"github.com/google/uuid"
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
	idxMgr, err := idx.NewManager(storage, dbname, collname)
	if err != nil {
		return nil, err
	}

	return &Collection{
		dbname:       dbname,
		collname:     collname,
		storage:      storage,
		initialized:  (idxMgr.DocumentCount() > 0),
		IndexManager: idxMgr,
	}, nil
}

// ensureInitialized ensures that the collection is initialized.
func (c *Collection) ensureInitialized() error {
	if c.initialized {
		return nil
	}

	if c.IndexManager.DocumentCount() > 0 {
		if err := c.IndexManager.SaveMetadata(); err != nil {
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
func (c *Collection) ensureDocumentID(doc map[string]any) string {
        val, ok := doc[consts.DocumentFieldID]

        if !ok {
                doc[consts.DocumentFieldID] = uuid.NewString()
        } else {
                tval := reflect.ValueOf(val)
                if tval.IsZero() && tval.Kind() == reflect.String {
                        doc[consts.DocumentFieldID] = uuid.NewString()
                }
        }

        return fmt.Sprintf("%v", doc[consts.DocumentFieldID])
}

// updateOne updates a single document by a filter.
func (c *Collection) updateOne(
	txn storage.Transaction,
	filter map[string]any,
	doc any,
	opts ...*options.UpdateOptions,
) UpdateOneResult {
	c.IndexManager.LoadMetadata()

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

        docUpdate := docMap
        if opt.Set == nil || !*opt.Set {
                for k, v := range result.Document() {
                        if _, ok := docUpdate[k]; !ok {
                                docUpdate[k] = v
                        }
                }
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

	err = c.IndexManager.IndexDocument(txn, docUpdate)
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
	c.IndexManager.LoadMetadata()

	// 1. Convertimos a BSON (map[string]interface{})
	parsed, err := bson.ConvertToMap(doc)
	if err != nil {
		return InsertOneResult{
			Err: fmt.Errorf("bson conversion failed: %w", err),
		}
	}

        // 2. Generamos ID único
        docID := c.ensureDocumentID(parsed)

        // 3. Verificamos unicidad en índices
        if err := c.IndexManager.CheckUniqueness(parsed); err != nil {
                return InsertOneResult{
                        Err: err,
                }
        }

        // 4. Serializamos a JSON
        data, err := bson.Marshal(parsed)
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

	c.IndexManager.IncrementDocumentCount()

	// 6. Registramos índices secundarios
        err = c.IndexManager.IndexDocument(txn, parsed)
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
	c.IndexManager.LoadMetadata()

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

	c.IndexManager.DecrementDocumentCount()

	err = c.IndexManager.DeleteDocumentIndexes(result.Document())
	if err != nil {
		return DeleteOneResult{
			Err: fmt.Errorf("delete document indexes failed: %w", err),
		}
	}

	c.IndexManager.SaveMetadata()

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

			if !aok || !bok || va == nil || vb == nil {
				continue
			}

			if result, ok := compareValues(va, vb); ok && result != 0 {
				if f.Order < 0 {
					return -result
				}

				return result
			}
		}

		return 0
	})
}

func compareValues(a, b any) (int, bool) {
	switch va := a.(type) {
	case string:
		if vb, ok := b.(string); ok {
			return cmp.Compare(strings.ToLower(va), strings.ToLower(vb)), true
		}
	case bool:
		if vb, ok := b.(bool); ok {
			ai, bi := 0, 0
			if va {
				ai = 1
			}
			if vb {
				bi = 1
			}

			return cmp.Compare(ai, bi), true
		}
	case time.Time:
		if vb, ok := b.(time.Time); ok {
			return cmp.Compare(va.UnixNano(), vb.UnixNano()), true
		}
	case time.Duration:
		if vb, ok := b.(time.Duration); ok {
			return cmp.Compare(va, vb), true
		}
	default:
		return compareNumeric(a, b)
	}

	return 0, false
}

func compareNumeric(a, b any) (int, bool) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	if av.Kind() != bv.Kind() {
		return 0, false
	}

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ai, bi := av.Int(), bv.Int()
		switch {
		case ai < bi:
			return -1, true
		case ai > bi:
			return 1, true
		}
		return 0, true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		au, bu := av.Uint(), bv.Uint()
		switch {
		case au < bu:
			return -1, true
		case au > bu:
			return 1, true
		}
		return 0, true
	case reflect.Float32, reflect.Float64:
		af, bf := av.Float(), bv.Float()
		switch {
		case af < bf:
			return -1, true
		case af > bf:
			return 1, true
		}
		return 0, true
	}

	return 0, false
}
