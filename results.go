package gopherdb

import (
	"fmt"
	"reflect"

	"github.com/NaujOyamat/gopherdb/internal/bson"
	"github.com/NaujOyamat/gopherdb/internal/storage"
)

// InsertOneResult es el resultado de una inserción.
type InsertOneResult struct {
	InsertedID any
	Err        error
}

// InsertManyResult es el resultado de una inserción de múltiples documentos.
type InsertManyResult struct {
	InsertedIDs []any
	Err         error
}

// FindOneResult es el resultado de una consulta de un documento.
type FindOneResult struct {
	raw       storage.KV
	IndexUsed *IndexModel
	Err       error
}

// Document returns the document of the find one result.
func (r *FindOneResult) Document() map[string]any {
	if r.Err != nil {
		return nil
	}

	return r.raw.Document()
}

// Unmarshal unmarshals the result into a struct.
func (r *FindOneResult) Unmarshal(result any) error {
	if r.Err != nil {
		return r.Err
	}

	return bson.Unmarshal(r.raw.Value, result)
}

// FindResult es el resultado de una consulta.
type FindResult struct {
	raw        []storage.KV
	TotalCount int64
	IndexUsed  *IndexModel
	Err        error
}

// Unmarshal unmarshals the results into a slice of the given type.
func (r *FindResult) Unmarshal(results any) error {
	if r.Err != nil {
		return r.Err
	}

	resultsVal := reflect.ValueOf(results)
	if resultsVal.Kind() != reflect.Ptr {
		return fmt.Errorf("results argument must be a pointer to a slice, but was a %s", resultsVal.Kind())
	}

	sliceVal := resultsVal.Elem()
	if sliceVal.Kind() == reflect.Interface {
		sliceVal = sliceVal.Elem()
	}

	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("results argument must be a pointer to a slice, but was a pointer to %s", sliceVal.Kind())
	}

	elemType := sliceVal.Type().Elem()
	index := 0

	for _, kv := range r.raw {
		if sliceVal.Len() == index {
			// slice is full
			newElem := reflect.New(elemType)
			sliceVal = reflect.Append(sliceVal, newElem.Elem())
			sliceVal = sliceVal.Slice(0, sliceVal.Cap())
		}

		currElem := sliceVal.Index(index).Addr().Interface()

		err := bson.Unmarshal(kv.Value, currElem)

		if err != nil {
			return fmt.Errorf("error unmarshalling result: %w", err)
		}

		index++
	}

	resultsVal.Elem().Set(sliceVal.Slice(0, index))

	return nil
}

// DeleteOneResult es el resultado de una eliminación de un documento.
type DeleteOneResult struct {
	DeletedID any
	Err       error
}

// DeleteManyResult es el resultado de una eliminación de múltiples documentos.
type DeleteManyResult struct {
	DeletedIDs []any
	Err        error
}

// UpdateOneResult es el resultado de una actualización de un documento.
type UpdateOneResult struct {
	UpsertedID any
	Err        error
}

// UpdateManyResult es el resultado de una actualización de múltiples documentos.
type UpdateManyResult struct {
	UpsertedIDs []any
	Err         error
}
