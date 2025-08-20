package gopherdb

import "context"

// CreateIndex crea un nuevo índice en la colección.
func (c *Collection) CreateIndex(ctx context.Context, index IndexModel) error {
	return c.IndexManager.CreateMany(ctx, []IndexModel{index})
}

// CreateManyIndexes crea múltiples índices en la colección.
func (c *Collection) CreateManyIndexes(ctx context.Context, indexes []IndexModel) error {
	return c.IndexManager.CreateMany(ctx, indexes)
}
