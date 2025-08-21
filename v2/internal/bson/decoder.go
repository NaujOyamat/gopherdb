package bson

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetDecoder returns a new decoder for the given data and options.
func GetDecoder(
	data []byte,
	opts *options.BSONOptions,
) (*bson.Decoder, error) {
	dec, err := bson.NewDecoder(bsonrw.NewBSONDocumentReader(data))
	if err != nil {
		return nil, err
	}

	if opts != nil {
		if opts.AllowTruncatingDoubles {
			dec.AllowTruncatingDoubles()
		}

		if opts.BinaryAsSlice {
			dec.BinaryAsSlice()
		}

		if opts.DefaultDocumentD {
			dec.DefaultDocumentD()
		}

		if opts.DefaultDocumentM {
			dec.DefaultDocumentM()
		}

		if opts.UseJSONStructTags {
			dec.UseJSONStructTags()
		}

		if opts.UseLocalTimeZone {
			dec.UseLocalTimeZone()
		}

		if opts.ZeroMaps {
			dec.ZeroMaps()
		}

		if opts.ZeroStructs {
			dec.ZeroStructs()
		}
	}

	return dec, nil
}
