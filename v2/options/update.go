package options

// Update crea una nueva instancia de updateOptions.
func Update() *UpdateOptions {
	not := false

	return &UpdateOptions{
		Upsert: &not,
		Set:    &not,
	}
}

// UpdateOptions es un struct que contiene las opciones para una actualización.
type UpdateOptions struct {
	Upsert *bool
	Set    *bool
}

// Merge combina las opciones de varias consultas.
func (o *UpdateOptions) Merge(opts ...*UpdateOptions) *UpdateOptions {
	for _, opt := range opts {
		if opt.Upsert != nil {
			o.Upsert = opt.Upsert
		}

		if opt.Set != nil {
			o.Set = opt.Set
		}
	}

	return o
}

// SetUpsert establece el valor de la opción Upsert.
func (o *UpdateOptions) SetUpsert(upsert bool) *UpdateOptions {
	o.Upsert = &upsert

	return o
}

// SetSet establece el valor de la opción Set.
func (o *UpdateOptions) SetSet(set bool) *UpdateOptions {
	o.Set = &set

	return o
}
