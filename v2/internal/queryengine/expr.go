package queryengine

import (
	"fmt"
)

// Expr is a query expression.
type Expr interface {
	Evaluate(doc map[string]any) bool
}

// AndExpr is a query expression that evaluates to true if all of the clauses are true.
type AndExpr struct {
	Clauses []Expr
}

// Evaluate evaluates the expression.
func (a AndExpr) Evaluate(doc map[string]any) bool {
	for _, clause := range a.Clauses {
		if !clause.Evaluate(doc) {
			return false
		}
	}

	return true
}

// OrExpr is a query expression that evaluates to true if any of the clauses are true.
type OrExpr struct {
	Clauses []Expr
}

// Evaluate evaluates the expression.
func (o OrExpr) Evaluate(doc map[string]any) bool {
	for _, clause := range o.Clauses {
		if clause.Evaluate(doc) {
			return true
		}
	}

	return false
}

// ComparisonExpr is a query expression that evaluates to true if the field matches the value.
type ComparisonExpr struct {
	Field    string
	Operator Operator
	Value    any
}

// Evaluate evaluates the expression.
func (c ComparisonExpr) Evaluate(doc map[string]any) bool {
	val, ok := doc[c.Field]
	if !ok {
		return false
	}

	switch c.Operator {
	case OperatorEqual:
		return compare(val, c.Value) == 0
	case OperatorNotEqual:
		return compare(val, c.Value) != 0
	case OperatorGreaterThan:
		return compare(val, c.Value) > 0
	case OperatorLessThan:
		return compare(val, c.Value) < 0
	case OperatorGreaterThanOrEqual:
		return compare(val, c.Value) >= 0
	case OperatorLessThanOrEqual:
		return compare(val, c.Value) <= 0
	case OperatorExists:
		exists := val != nil

		return exists == c.Value
	case OperatorIn:
		list, ok := c.Value.([]any)
		if !ok {
			return false
		}

		for _, item := range list {
			if compare(val, item) == 0 {
				return true
			}
		}

		return false
	default:
		return false
	}
}

// compare compares two values.
func compare(a, b any) int {
	af, aok := toFloat64(a)
	bf, bok := toFloat64(b)

	if aok && bok {
		if af < bf {
			return -1
		} else if af > bf {
			return 1
		}

		return 0
	}

	return 0
}

// toFloat64 converts a value to a float64.
func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case float64:
		return n, true
	case float32:
		return float64(n), true
	default:
		return 0, false
	}
}

// ParseFilter parses a filter map into an Expr.
func ParseFilter(filter map[string]any) (Expr, error) {
	clauses := []Expr{}

	for key, value := range filter {
		switch key {
		case OperatorAnd.String():
			arr, ok := value.([]any)
			if !ok {
				return nil, fmt.Errorf("$and must be array")
			}

			sub := []Expr{}

			for _, item := range arr {
				fmap, ok := item.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid $and clause")
				}

				expr, err := ParseFilter(fmap)
				if err != nil {
					return nil, err
				}

				sub = append(sub, expr)
			}

			clauses = append(clauses, AndExpr{Clauses: sub})
		case OperatorOr.String():
			arr, ok := value.([]any)
			if !ok {
				return nil, fmt.Errorf("$or must be array")
			}

			sub := []Expr{}

			for _, item := range arr {
				fmap, ok := item.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid $or clause")
				}

				expr, err := ParseFilter(fmap)
				if err != nil {
					return nil, err
				}

				sub = append(sub, expr)
			}

			clauses = append(clauses, OrExpr{Clauses: sub})
		default:
			switch typed := value.(type) {
			case map[string]any:
				for op, val := range typed {
					clauses = append(clauses, ComparisonExpr{
						Field:    key,
						Operator: Operator(op),
						Value:    val,
					})
				}
			default:
				clauses = append(clauses, ComparisonExpr{
					Field:    key,
					Operator: OperatorEqual,
					Value:    typed,
				})
			}
		}
	}

	if len(clauses) == 1 {
		return clauses[0], nil
	}

	return AndExpr{Clauses: clauses}, nil
}
