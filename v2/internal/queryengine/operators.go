package queryengine

type Operator string

const (
	// OperatorAnd is the and operator.
	OperatorAnd Operator = "$and"
	// OperatorOr is the or operator.
	OperatorOr Operator = "$or"
	// OperatorEqual is the equal operator.
	OperatorEqual Operator = "$eq"
	// OperatorNotEqual is the not equal operator.
	OperatorNotEqual Operator = "$ne"
	// OperatorGreaterThan is the greater than operator.
	OperatorGreaterThan Operator = "$gt"
	// OperatorLessThan is the less than operator.
	OperatorLessThan Operator = "$lt"
	// OperatorGreaterThanOrEqual is the greater than or equal operator.
	OperatorGreaterThanOrEqual Operator = "$gte"
	// OperatorLessThanOrEqual is the less than or equal operator.
	OperatorLessThanOrEqual Operator = "$lte"
	// OperatorIn is the in operator.
	OperatorIn Operator = "$in"
	// OperatorExists is the exists operator.
	OperatorExists Operator = "$exists"
	// OperatorType is the type operator.
	OperatorType Operator = "$type"
)

func (o Operator) String() string {
	return string(o)
}

func (o Operator) IsValid() bool {
	switch o {
	case OperatorEqual, OperatorNotEqual, OperatorGreaterThan,
		OperatorLessThan, OperatorGreaterThanOrEqual, OperatorLessThanOrEqual,
		OperatorIn, OperatorExists, OperatorType:
		return true
	default:
		return false
	}
}
