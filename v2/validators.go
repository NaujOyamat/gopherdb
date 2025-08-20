package gopherdb

import "reflect"

// validateDocumentSliceType valida que el tipo de un slice de documentos sea válido.
func validateDocumentSliceType(docs any) (reflect.Value, error) {
        if docs == nil {
                return reflect.Value{}, ErrDocumentSliceIsNil
        }

        switch v := docs.(type) {
        case []map[string]any:
                if len(v) == 0 {
                        return reflect.Value{}, ErrDocumentSliceEmpty
                }
                return reflect.ValueOf(v), nil
        default:
                docsVal := reflect.ValueOf(docs)

                if docsVal.Kind() == reflect.Ptr {
                        if docsVal.IsNil() {
                                return reflect.Value{}, ErrDocumentSlicePointerIsNil
                        }

                        docsVal = docsVal.Elem()
                }

                if docsVal.Kind() == reflect.Interface {
                        docsVal = docsVal.Elem()
                }

                if docsVal.Kind() != reflect.Slice {
                        return reflect.Value{}, ErrDocumentSliceTypeInvalid
                }

                if docsVal.Len() == 0 {
                        return reflect.Value{}, ErrDocumentSliceEmpty
                }

                elemType := docsVal.Type().Elem()
                if elemType.Kind() == reflect.Ptr {
                        elemType = elemType.Elem()
                }

                if elemType.Kind() != reflect.Map && elemType.Kind() != reflect.Struct {
                        return reflect.Value{}, ErrDocumentSliceElementTypeInvalid
                }

                if elemType.Kind() == reflect.Map {
                        if elemType.Key().Kind() != reflect.String || elemType.Elem().Kind() != reflect.Interface {
                                return reflect.Value{}, ErrDocumentSliceElementTypeInvalid
                        }
                }

                return docsVal, nil
        }
}

// validateDocumentType valida que el tipo de un documento sea válido.
func validateDocumentType(doc any) error {
        if doc == nil {
                return ErrDocumentIsNil
        }

        switch v := doc.(type) {
        case map[string]any:
                return nil
        case *map[string]any:
                if v == nil {
                        return ErrDocumentPointerIsNil
                }
                return nil
        default:
                docVal := reflect.ValueOf(doc)
                if docVal.Kind() == reflect.Ptr {
                        if docVal.IsNil() {
                                return ErrDocumentPointerIsNil
                        }
                        docVal = docVal.Elem()
                }

                if docVal.Kind() != reflect.Struct {
                        return ErrDocumentTypeInvalid
                }

                return nil
        }
}
