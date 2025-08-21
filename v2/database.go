package gopherdb

import "github.com/NaujOyamat/gopherdb/v2/internal/storage"

// Database es una base de datos.
type Database struct {
	name    string
	colls   []*Collection
	storage storage.Storage
}

// NewDatabase crea una nueva instancia de Database.
func NewDatabase(name, path string) (*Database, error) {
	engine, err := storage.NewStorage(path)
	if err != nil {
		return nil, err
	}

	return &Database{
		name:    name,
		storage: engine,
	}, nil
}

// Collection devuelve una instancia de Collection para la base de datos
func (db *Database) Collection(name string) (*Collection, error) {
	col, err := newCollection(db.storage, db.name, name)
	if err != nil {
		return nil, err
	}

	db.colls = append(db.colls, col)

	return col, nil
}

// Close cierra la base de datos
func (db *Database) Close() error {
	return db.storage.Close()
}
