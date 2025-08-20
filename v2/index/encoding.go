package index

import (
        "fmt"
        "net/url"
        "time"
)

// encodeForLexOrder encodes a value for lexicographical order.
func encodeForLexOrder(val any, desc bool) (string, error) {
        var s string

        switch v := val.(type) {
        case string:
                s = v
        case int:
                s = fmt.Sprintf("%020d", v)
        case int8:
                s = fmt.Sprintf("%020d", v)
        case int16:
                s = fmt.Sprintf("%020d", v)
        case int32:
                s = fmt.Sprintf("%020d", v)
        case int64:
                s = fmt.Sprintf("%020d", v)
        case uint:
                s = fmt.Sprintf("%020d", v)
        case uint8:
                s = fmt.Sprintf("%020d", v)
        case uint16:
                s = fmt.Sprintf("%020d", v)
        case uint32:
                s = fmt.Sprintf("%020d", v)
        case uint64:
                s = fmt.Sprintf("%020d", v)
        case float32:
                s = fmt.Sprintf("%020.6f", v)
        case float64:
                s = fmt.Sprintf("%020.6f", v)
        case bool:
                if v {
                        s = "1"
                } else {
                        s = "0"
                }
        case time.Time:
                s = v.Format(time.RFC3339)
        case []byte:
                s = string(v)
        case time.Duration:
                s = fmt.Sprintf("%020d", v)
        default:
                return "", fmt.Errorf("%w: %T", ErrUnsupportedIndexValue, v)
        }

        s = url.PathEscape(s)

        if desc {
                s = invertLex(s)
        }

        return s, nil
}

// invertLex inverts the lexicographical order of a string.
func invertLex(s string) string {
	b := []byte(s)
	for i := range b {
		b[i] = 255 - b[i]
	}

	return fmt.Sprintf("~%s", b)
}
