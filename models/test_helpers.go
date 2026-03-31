package models

import (
	"github.com/shopspring/decimal"
)

// mustDecimal is a test helper that converts a string to Decimal, panicking on error.
// Used across model tests for convenient decimal initialization.
func mustDecimal(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic("invalid decimal: " + s)
	}
	return d
}
