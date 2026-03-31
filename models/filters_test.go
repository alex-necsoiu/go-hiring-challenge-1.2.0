package models

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestProductFilterCreation(t *testing.T) {
	tests := []struct {
		name     string
		filter   ProductFilter
		validate func(t *testing.T, pf ProductFilter)
	}{
		{
			name: "filter with all fields set",
			filter: ProductFilter{
				Offset:       10,
				Limit:        25,
				CategoryCode: "CLOTHING",
				MaxPrice:     mustDecimal("99.99"),
			},
			validate: func(t *testing.T, pf ProductFilter) {
				assert.Equal(t, 10, pf.Offset)
				assert.Equal(t, 25, pf.Limit)
				assert.Equal(t, "CLOTHING", pf.CategoryCode)
				assert.True(t, pf.MaxPrice.Equal(mustDecimal("99.99")))
			},
		},
		{
			name: "filter with no category or price",
			filter: ProductFilter{
				Offset: 0,
				Limit:  50,
			},
			validate: func(t *testing.T, pf ProductFilter) {
				assert.Equal(t, 0, pf.Offset)
				assert.Equal(t, 50, pf.Limit)
				assert.Equal(t, "", pf.CategoryCode)
				assert.True(t, pf.MaxPrice.IsZero())
			},
		},
		{
			name: "filter with only category",
			filter: ProductFilter{
				CategoryCode: "SHOES",
			},
			validate: func(t *testing.T, pf ProductFilter) {
				assert.Equal(t, 0, pf.Offset)
				assert.Equal(t, 0, pf.Limit)
				assert.Equal(t, "SHOES", pf.CategoryCode)
				assert.True(t, pf.MaxPrice.IsZero())
			},
		},
		{
			name: "filter with only price",
			filter: ProductFilter{
				MaxPrice: mustDecimal("50.00"),
			},
			validate: func(t *testing.T, pf ProductFilter) {
				assert.Equal(t, 0, pf.Offset)
				assert.Equal(t, 0, pf.Limit)
				assert.Equal(t, "", pf.CategoryCode)
				assert.True(t, pf.MaxPrice.Equal(mustDecimal("50.00")))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.filter)
		})
	}
}

func TestProductFilterFieldTypes(t *testing.T) {
	pf := ProductFilter{
		Offset:       5,
		Limit:        20,
		CategoryCode: "TYPE-CHECK",
		MaxPrice:     mustDecimal("123.45"),
	}

	// Verify field types through assignments
	assertInt := func(i int) bool { return true }
	assertString := func(s string) bool { return true }
	assertDecimal := func(d decimal.Decimal) bool { return true }

	assert.True(t, assertInt(pf.Offset))
	assert.True(t, assertInt(pf.Limit))
	assert.True(t, assertString(pf.CategoryCode))
	assert.True(t, assertDecimal(pf.MaxPrice))
}

func TestProductFilterZeroValue(t *testing.T) {
	pf := ProductFilter{}

	assert.Equal(t, 0, pf.Offset)
	assert.Equal(t, 0, pf.Limit)
	assert.Equal(t, "", pf.CategoryCode)
	assert.True(t, pf.MaxPrice.IsZero())
}

func TestProductFilterPagination(t *testing.T) {
	tests := []struct {
		name   string
		offset int
		limit  int
	}{
		{
			name:   "first page",
			offset: 0,
			limit:  10,
		},
		{
			name:   "second page",
			offset: 10,
			limit:  10,
		},
		{
			name:   "large offset",
			offset: 1000,
			limit:  50,
		},
		{
			name:   "single item",
			offset: 0,
			limit:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := ProductFilter{
				Offset: tt.offset,
				Limit:  tt.limit,
			}
			assert.Equal(t, tt.offset, pf.Offset)
			assert.Equal(t, tt.limit, pf.Limit)
		})
	}
}

func TestProductFilterPriceFiltering(t *testing.T) {
	tests := []struct {
		name     string
		maxPrice decimal.Decimal
		expected string
	}{
		{
			name:     "small price",
			maxPrice: mustDecimal("9.99"),
			expected: "9.99",
		},
		{
			name:     "medium price",
			maxPrice: mustDecimal("99.99"),
			expected: "99.99",
		},
		{
			name:     "large price",
			maxPrice: mustDecimal("9999.99"),
			expected: "9999.99",
		},
		{
			name:     "zero price (no filtering)",
			maxPrice: decimal.Zero,
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := ProductFilter{MaxPrice: tt.maxPrice}
			assert.Equal(t, tt.expected, pf.MaxPrice.String())
		})
	}
}

func TestProductFilterCategoryFiltering(t *testing.T) {
	tests := []struct {
		name         string
		categoryCode string
	}{
		{
			name:         "clothing category",
			categoryCode: "CLOTHING",
		},
		{
			name:         "shoes category",
			categoryCode: "SHOES",
		},
		{
			name:         "empty category (no filtering)",
			categoryCode: "",
		},
		{
			name:         "category with special chars",
			categoryCode: "SPECIAL-CATS_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := ProductFilter{CategoryCode: tt.categoryCode}
			assert.Equal(t, tt.categoryCode, pf.CategoryCode)
		})
	}
}

func TestProductFilterCombinations(t *testing.T) {
	// Test various combinations of filter parameters
	testCases := []struct {
		name     string
		offset   int
		limit    int
		category string
		maxPrice string
	}{
		{"all filters", 20, 10, "CLOTHING", "100"},
		{"pagination only", 50, 25, "", "0"},
		{"category and price", 0, 10, "SHOES", "75.5"},
		{"price only", 0, 0, "", "50"},
		{"no filters", 0, 0, "", "0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pf := ProductFilter{
				Offset:       tc.offset,
				Limit:        tc.limit,
				CategoryCode: tc.category,
				MaxPrice:     mustDecimal(tc.maxPrice),
			}

			assert.Equal(t, tc.offset, pf.Offset)
			assert.Equal(t, tc.limit, pf.Limit)
			assert.Equal(t, tc.category, pf.CategoryCode)
			assert.Equal(t, tc.maxPrice, pf.MaxPrice.String())
		})
	}
}
