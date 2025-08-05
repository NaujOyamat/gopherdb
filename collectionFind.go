package gopherdb

import (
	"fmt"

	"github.com/NaujOyamat/gopherdb/internal/queryengine"
	"github.com/NaujOyamat/gopherdb/internal/storage"
	"github.com/NaujOyamat/gopherdb/options"
)

// FindByID finds a document by its ID.
func (c *Collection) FindByID(id any) FindOneResult {
	key := c.buildDocumentKey(fmt.Sprintf("%v", id))
	data, err := c.storage.Get(key)

	if err != nil {
		return FindOneResult{
			Err: fmt.Errorf("document not found: %w", err),
		}
	}

	return FindOneResult{
		raw: storage.KV{
			Key:   key,
			Value: data,
		},
	}
}

// FindOne finds a single document by a filter.
func (c *Collection) FindOne(
	filter map[string]any,
) FindOneResult {
	opts := options.Find().SetLimit(1)

	result := c.Find(filter, opts)
	if result.Err != nil {
		return FindOneResult{
			Err: result.Err,
		}
	}

	if len(result.raw) == 0 {
		return FindOneResult{
			Err: ErrDocumentNotFound,
		}
	}

	resultFind := FindOneResult{}

	if len(result.raw) > 0 {
		resultFind.raw = result.raw[0]
	}

	return resultFind
}

// Find finds documents by a filter.
func (c *Collection) Find(
	filter map[string]any,
	opts ...*options.FindOptions,
) FindResult {
	c.IndexManager.loadMetadata()

	opt := options.Find()
	if len(opts) > 0 {
		opt = opt.Merge(opts...)
	}

	planner := NewQueryPlanner(c.IndexManager.metadata.Indexes)
	plan := planner.Plan(filter, opt.Sort)

	expr, err := queryengine.ParseFilter(filter)
	if err != nil {
		return FindResult{
			Err: fmt.Errorf("invalid filter: %w", err),
		}
	}

	raw := make([]storage.KV, 0)
	totalCount := int64(0)

	if plan.IndexUsed != nil {
		docKeys := make([]string, 0)

		if plan.IndexFilter != nil && len(plan.IndexFilter) > 0 {
			docKeys, err = c.IndexManager.getDocumentIndexKeysByIndexAndFilter(*plan.IndexUsed, plan.IndexFilter)

			if err != nil {
				return FindResult{
					Err: fmt.Errorf("get document keys by index failed: %w", err),
				}
			}
		} else {
			docKeys, err = c.IndexManager.getDocumentIndexKeysByIndex(*plan.IndexUsed)
			if err != nil {
				return FindResult{
					Err: fmt.Errorf("get all document keys by index failed: %w", err),
				}
			}
		}

		if plan.UsedForSort {
			// ✅ Aplica paginación directamente
			start := 0
			end := len(docKeys)

			if opt.Skip != nil {
				start = int(*opt.Skip)
			}

			if opt.Limit != nil && start+int(*opt.Limit) < end {
				end = start + int(*opt.Limit)
			}

			if start < len(docKeys) {
				docKeys = docKeys[start:end]
			} else {
				docKeys = nil
			}
		}

		for _, docKey := range docKeys {
			k, err := c.IndexManager.getDocumentIdFromIndexKey(docKey)
			if err != nil {
				return FindResult{
					Err: fmt.Errorf("get document id from index key failed: %w", err),
				}
			}

			result := c.FindByID(k)
			if result.Err != nil && result.Err != ErrDocumentNotFound {
				return FindResult{
					Err: fmt.Errorf("get document by key failed: %w", result.Err),
				}
			}

			if expr.Evaluate(result.Document()) {
				raw = append(raw, result.raw)
				totalCount++
			}
		}
	} else {
		documentsKey := c.IndexManager.buildDocumentsKey()

		docs, err := c.storage.Scan(documentsKey)
		if err != nil {
			return FindResult{
				Err: fmt.Errorf("scan keys failed: %w", err),
			}
		}

		for _, kv := range docs {
			if expr.Evaluate(kv.Document()) {
				raw = append(raw, kv)
				totalCount++
			}
		}
	}

	result := FindResult{
		raw:        raw,
		IndexUsed:  plan.IndexUsed,
		TotalCount: totalCount,
	}

	if opt.Sort != nil && !plan.UsedForSort {
		c.sortDocuments(result.raw, opt)
	}

	if opt.Skip != nil {
		if int(*opt.Skip) < len(result.raw) {
			result.raw = result.raw[*opt.Skip:]
		} else {
			result.raw = nil
		}
	}

	if opt.Limit != nil {
		if int(*opt.Limit) < len(result.raw) {
			result.raw = result.raw[:*opt.Limit]
		}
	}

	return result
}
