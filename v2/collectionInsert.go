package gopherdb

import (
	"fmt"
)

// InsertOne inserts a single document into the collection.
func (c *Collection) InsertOne(doc any) InsertOneResult {
	_, err := validateDocumentType(doc)
	if err != nil {
		return InsertOneResult{
			Err: err,
		}
	}

	txn := c.storage.BeginTx()
	result := c.insertOne(txn, doc)

	if result.Err != nil {
		txn.Rollback()

		return result
	}

	if err := txn.Commit(); err != nil {
		return InsertOneResult{
			Err: fmt.Errorf("commit failed: %w", err),
		}
	}

	return result
}

// Insert inserts multiple documents into the collection.
func (c *Collection) Insert(docs any) InsertManyResult {
	resultsVal, err := validateDocumentSliceType(docs)
	if err != nil {
		return InsertManyResult{
			Err: err,
		}
	}

	insertedIDs := make([]any, 0)
	totalDocs := resultsVal.Len()
	batchSize := 100

	for i := 0; i < totalDocs; i += batchSize {
		end := min(i+batchSize, totalDocs)

		txn := c.storage.BeginTx()

		for j := i; j < end; j++ {
			doc := resultsVal.Index(j).Interface()
			result := c.insertOne(txn, doc)

			if result.Err != nil {
				txn.Rollback()

				return InsertManyResult{
					Err: result.Err,
				}
			}

			insertedIDs = append(insertedIDs, result.InsertedID)
		}

		if err := txn.Commit(); err != nil {
			return InsertManyResult{
				Err: fmt.Errorf("commit failed: %w", err),
			}
		}
	}

	return InsertManyResult{
		InsertedIDs: insertedIDs,
	}
}
