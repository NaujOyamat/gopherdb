package bson

import "go.mongodb.org/mongo-driver/bson"

// ConvertToMap converts a value to a map.
func ConvertToMap(o any) (map[string]any, error) {
	jsonData, err := bson.Marshal(o)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	err = bson.Unmarshal(jsonData, &m)

	if err != nil {
		return nil, err
	}

	return m, nil
}

// ConvertToStruct converts a map to a struct.
func ConvertToStruct(m map[string]any, o any) error {
	jsonData, err := bson.Marshal(m)
	if err != nil {
		return err
	}

	return bson.Unmarshal(jsonData, o)
}
