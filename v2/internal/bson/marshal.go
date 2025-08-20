package bson

import (
	"encoding/binary"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

// Marshal marshals a value to BSON.
func Marshal(v any) ([]byte, error) {
	return bson.Marshal(v)
}

// Unmarshal unmarshals a BSON value to a value.
func Unmarshal(data []byte, v any) error {
	return bson.Unmarshal(data, v)
}

// ValidateBSON validates a BSON document.
func ValidateBSON(data []byte) error {
	if len(data) < 5 {
		return fmt.Errorf("document too short")
	}

	expectedLen := int(binary.LittleEndian.Uint32(data[:4]))
	actualLen := len(data)

	if expectedLen != actualLen {
		return fmt.Errorf("length mismatch: declared %d, actual %d", expectedLen, actualLen)
	}

	if data[len(data)-1] != 0x00 {
		return fmt.Errorf("document missing 0x00 terminator")
	}

	var doc map[string]any
	if err := bson.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("bson unmarshal failed: %w", err)
	}

	return nil
}
