package gopherdb

import (
	"fmt"

	"github.com/NaujOyamat/gopherdb/v2/options"
)

// UpdateOne updates a single document by a filter.
func (c *Collection) UpdateOne(
	filter map[string]any,
	doc any,
	opts ...*options.UpdateOptions,
) UpdateOneResult {
        if err := validateDocumentType(doc); err != nil {
                return UpdateOneResult{
                        Err: err,
                }
        }

	txn := c.storage.BeginTx()
	result := c.updateOne(txn, filter, doc, opts...)

	if result.Err != nil {
		txn.Rollback()

		return result
	}

	if err := txn.Commit(); err != nil {
		return UpdateOneResult{
			Err: fmt.Errorf("commit failed: %w", err),
		}
	}

	return result
}

// Update updates multiple documents by a filter.
func (c *Collection) Update(
	filter map[string]any,
	docs any,
	opts ...*options.UpdateOptions,
) UpdateManyResult {
	resultsVal, err := validateDocumentSliceType(docs)
	if err != nil {
		return UpdateManyResult{
			Err: err,
		}
	}

	upsertedIDs := make([]any, 0)

	txn := c.storage.BeginTx()

	for i := range resultsVal.Len() {
		doc := resultsVal.Index(i).Interface()
		result := c.updateOne(txn, filter, doc, opts...)

		if result.Err != nil {
			txn.Rollback()

			return UpdateManyResult{
				Err: fmt.Errorf("update one failed: %w", result.Err),
			}
		}

		if result.UpsertedID != nil {
			upsertedIDs = append(upsertedIDs, result.UpsertedID)
		}
	}

	if err := txn.Commit(); err != nil {
		return UpdateManyResult{
			Err: fmt.Errorf("commit failed: %w", err),
		}
	}

	return UpdateManyResult{
		UpsertedIDs: upsertedIDs,
	}
}
