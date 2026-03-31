package models

import (
	"github.com/shopspring/decimal"
)

// ProductFilter contains pagination and filtering options for product queries.
type ProductFilter struct {
	Offset       int
	Limit        int
	CategoryCode string
	MaxPrice     decimal.Decimal
}
