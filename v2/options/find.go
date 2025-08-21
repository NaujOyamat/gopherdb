package options

// SortField es un struct que contiene el campo y el orden para una consulta.
type SortField struct {
	Field string
	Order int
}

// FindOptions es un struct que contiene las opciones para una consulta.
type FindOptions struct {
	Skip  *int64
	Limit *int64
	Sort  []SortField
}

// Find crea una nueva instancia de findOptions.
func Find() *FindOptions {
	return &FindOptions{}
}

// Merge combina las opciones de varias consultas.
func (o *FindOptions) Merge(opts ...*FindOptions) *FindOptions {
	for _, opt := range opts {
		if opt.Skip != nil {
			o.Skip = opt.Skip
		}

		if opt.Limit != nil {
			o.Limit = opt.Limit
		}

		if opt.Sort != nil {
			o.Sort = opt.Sort
		}
	}

	return o
}

// SetSkip establece el número de documentos a saltar.
func (o *FindOptions) SetSkip(skip int64) *FindOptions {
	o.Skip = &skip

	return o
}

// SetLimit establece el número de documentos a devolver.
func (o *FindOptions) SetLimit(limit int64) *FindOptions {
	o.Limit = &limit

	return o
}

// SetSort establece el campo y el orden para una consulta.
func (o *FindOptions) SetSort(sort SortField) *FindOptions {
	if o.Sort == nil {
		o.Sort = make([]SortField, 0)
	}

	o.Sort = append(o.Sort, sort)

	return o
}
