package models

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestProductTableName(t *testing.T) {
	p := &Product{}
	assert.Equal(t, "products", p.TableName())
}

func TestProductCreation(t *testing.T) {
	tests := []struct {
		name     string
		product  Product
		validate func(t *testing.T, p Product)
	}{
		{
			name: "complete product with all fields",
			product: Product{
				ID:         1,
				Code:       "PROD001",
				Name:       "Test Product",
				Price:      mustDecimal("99.99"),
				CategoryID: 5,
				Category: Category{
					ID:   5,
					Code: "TEST-CAT",
					Name: "Test Category",
				},
				Variants: []Variant{
					{
						ID:        1,
						ProductID: 1,
						Name:      "Size M",
						SKU:       "SKU-001",
						Price:     mustDecimal("50.00"),
					},
				},
			},
			validate: func(t *testing.T, p Product) {
				assert.Equal(t, uint(1), p.ID)
				assert.Equal(t, "PROD001", p.Code)
				assert.Equal(t, "Test Product", p.Name)
				assert.True(t, p.Price.Equal(mustDecimal("99.99")))
				assert.Equal(t, uint(5), p.CategoryID)
				assert.Equal(t, "Test Category", p.Category.Name)
				assert.Len(t, p.Variants, 1)
				assert.Equal(t, "Size M", p.Variants[0].Name)
			},
		},
		{
			name: "minimal product with zero values",
			product: Product{
				Code:  "MIN-001",
				Name:  "Minimal",
				Price: decimal.Zero,
			},
			validate: func(t *testing.T, p Product) {
				assert.Equal(t, uint(0), p.ID)
				assert.Equal(t, "MIN-001", p.Code)
				assert.Equal(t, "Minimal", p.Name)
				assert.True(t, p.Price.IsZero())
				assert.Equal(t, uint(0), p.CategoryID)
			},
		},
		{
			name: "product with decimal price",
			product: Product{
				Code:  "DEC-001",
				Name:  "Decimal Product",
				Price: mustDecimal("123.45"),
			},
			validate: func(t *testing.T, p Product) {
				expected := mustDecimal("123.45")
				assert.True(t, p.Price.Equal(expected))
				assert.Equal(t, "123.45", p.Price.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.product)
		})
	}
}

func TestProductFieldTypes(t *testing.T) {
	p := Product{
		ID:         42,
		Code:       "TYPE-TEST",
		Name:       "Field Type Test",
		Price:      mustDecimal("999.99"),
		CategoryID: 10,
	}

	// Verify field types through assignments
	assertUint := func(v uint) bool { return true }
	assertString := func(v string) bool { return true }
	assertDecimal := func(v decimal.Decimal) bool { return true }

	assert.True(t, assertUint(p.ID))
	assert.True(t, assertString(p.Code))
	assert.True(t, assertString(p.Name))
	assert.True(t, assertDecimal(p.Price))
	assert.True(t, assertUint(p.CategoryID))
}

func TestProductZeroValue(t *testing.T) {
	p := Product{}

	assert.Equal(t, uint(0), p.ID)
	assert.Equal(t, "", p.Code)
	assert.Equal(t, "", p.Name)
	assert.True(t, p.Price.IsZero())
	assert.Equal(t, uint(0), p.CategoryID)
	assert.Equal(t, Category{}, p.Category)
	assert.Len(t, p.Variants, 0)
}

func TestProductPriceHandling(t *testing.T) {
	tests := []struct {
		name     string
		price    decimal.Decimal
		expected string
	}{
		{
			name:     "zero price",
			price:    decimal.Zero,
			expected: "0",
		},
		{
			name:     "positive price",
			price:    mustDecimal("49.99"),
			expected: "49.99",
		},
		{
			name:     "high precision price",
			price:    mustDecimal("1234567.89"),
			expected: "1234567.89",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Product{Price: tt.price}
			assert.True(t, p.Price.Equal(mustDecimal(tt.expected)))
		})
	}
}
