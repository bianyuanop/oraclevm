package oracle

import "errors"

var (
	ErrEmptyEntities              = errors.New("Aggregating empty entities")
	ErrNoAggregationRule          = errors.New("No such aggregation rule")
	ErrZeroDenominator            = errors.New("Deviding zero")
	ErrOutOfEntityCollectionRange = errors.New("Entity collection Id out of range")
)
