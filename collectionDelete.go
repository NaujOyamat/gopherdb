package gopherdb

import (
	"fmt"

	"github.com/NaujOyamat/gopherdb/internal/consts"
)

// DeleteOne deletes a single document by a filter.
func (c *Collection) DeleteOne(filter map[string]any) DeleteOneResult {
	txn := c.storage.BeginTx()
	result := c.deleteOne(txn, filter)

	if result.Err != nil {
		txn.Rollback()

		return result
	}

	if err := txn.Commit(); err != nil {
		return DeleteOneResult{
			Err: fmt.Errorf("commit failed: %w", err),
		}
	}

	return result
}

// DeleteByID deletes a single document by its ID.
func (c *Collection) DeleteByID(id any) DeleteOneResult {
	txn := c.storage.BeginTx()
	result := c.deleteOne(txn, map[string]any{"_id": id})

	if result.Err != nil {
		txn.Rollback()

		return result
	}

	if err := txn.Commit(); err != nil {
		return DeleteOneResult{
			Err: fmt.Errorf("commit failed: %w", err),
		}
	}

	return result
}

// Delete deletes multiple documents by a filter.
func (c *Collection) Delete(filter map[string]any) DeleteManyResult {
	results := c.Find(filter)

	if results.Err != nil {
		return DeleteManyResult{
			Err: results.Err,
		}
	}

	txn := c.storage.BeginTx()

	deletedIDs := make([]any, 0)

	for _, kv := range results.raw {
		if err := txn.Delete(kv.Key); err != nil {
			txn.Rollback()

			return DeleteManyResult{
				Err: fmt.Errorf("delete failed: %w", err),
			}
		}

		match, err := consts.DocumentKeyPathmatcher.Match(kv.Key)
		if err != nil {
			txn.Rollback()

			return DeleteManyResult{
				Err: fmt.Errorf("match failed: %w", err),
			}
		}

		deletedIDs = append(deletedIDs, match["docId"])
	}

	if err := txn.Commit(); err != nil {
		return DeleteManyResult{
			Err: fmt.Errorf("commit failed: %w", err),
		}
	}

	return DeleteManyResult{
		DeletedIDs: deletedIDs,
	}
}
