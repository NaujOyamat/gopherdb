package gopherdb

import (
	"github.com/NaujOyamat/gopherdb/internal/queryengine"
	"github.com/NaujOyamat/gopherdb/options"
)

// QueryPlan is the plan for a query.
type QueryPlan struct {
	IndexUsed   *IndexModel
	IndexFilter map[string]any
	UsedForSort bool
	IsExact     bool
}

// QueryPlanner is the planner for a query.
type QueryPlanner struct {
	indexes []IndexModel
}

// NewQueryPlanner creates a new QueryPlanner.
func NewQueryPlanner(indexes []IndexModel) *QueryPlanner {
	return &QueryPlanner{
		indexes: indexes,
	}
}

// Plan plans a query.
func (qp *QueryPlanner) Plan(
	filter map[string]any,
	sort []options.SortField,
) *QueryPlan {
	flatFilter := qp.flattenFilter(filter)

	var best *IndexModel

	var bestFilter map[string]any

	maxFilterMatch := 0
	bestUsedForSort := false

checkIndex:
	for _, index := range qp.indexes {
		localFilter := map[string]any{}
		matchCount := 0

		// Evaluamos coincidencias en el filtro
		for i, field := range index.Fields {
			val, ok := flatFilter[field.Name]
			if !ok {
				break
			}

			opMap, ok := val.(map[string]any)
			if !ok {
				break
			}

			if eqVal, ok := opMap[queryengine.OperatorEqual.String()]; ok {
				localFilter[field.Name] = eqVal
				matchCount++
			} else if neVal, ok := opMap[queryengine.OperatorNotEqual.String()]; ok {
				localFilter[field.Name] = map[string]any{queryengine.OperatorNotEqual.String(): neVal}
				matchCount++
			} else {
				break
			}

			if matchCount != i+1 {
				break
			}
		}

		// Si no hay filtro, evaluamos si el índice calza con el orden
		if matchCount == 0 && len(sort) > 0 {
			if qp.indexSupportsSort(index, sort) {
				best = &index
				bestFilter = map[string]any{}
				bestUsedForSort = true

				break checkIndex
			}
		}

		if matchCount > maxFilterMatch {
			best = &index
			bestFilter = localFilter
			maxFilterMatch = matchCount
			bestUsedForSort = qp.indexSupportsSort(index, sort)
		}
	}

	return &QueryPlan{
		IndexUsed:   best,
		IndexFilter: bestFilter,
		UsedForSort: bestUsedForSort,
		IsExact:     best != nil && len(bestFilter) == len(best.Fields),
	}
}

// flattenFilter flattens a filter.
func (qp *QueryPlanner) flattenFilter(filter map[string]any) map[string]any {
	out := map[string]any{}

	for k, v := range filter {
		switch val := v.(type) {
		case map[string]any:
			out[k] = val
		default:
			out[k] = map[string]any{queryengine.OperatorEqual.String(): val}
		}
	}

	return out
}

func (qp *QueryPlanner) indexSupportsSort(index IndexModel, sort []options.SortField) bool {
	if len(sort) > len(index.Fields) {
		return false
	}

	for i, sf := range sort {
		if index.Fields[i].Name != sf.Field {
			return false
		}

		if index.Fields[i].Order != sf.Order {
			return false
		}
	}

	return true
}
