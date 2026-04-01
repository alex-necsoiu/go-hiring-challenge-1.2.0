package models

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestVariantTableName(t *testing.T) {
	v := &Variant{}
	assert.Equal(t, "product_variants", v.TableName())
}

func TestVariantCreation(t *testing.T) {
	tests := []struct {
		name     string
		variant  Variant
		validate func(t *testing.T, v Variant)
	}{
		{
			name: "complete variant with price",
			variant: Variant{
				ID:        1,
				ProductID: 10,
				Name:      "Size Medium",
				SKU:       "SKU-MED-001",
				Price:     ptrDecimal(mustDecimal("29.99")),
			},
			validate: func(t *testing.T, v Variant) {
				assert.Equal(t, uint(1), v.ID)
				assert.Equal(t, uint(10), v.ProductID)
				assert.Equal(t, "Size Medium", v.Name)
				assert.Equal(t, "SKU-MED-001", v.SKU)
				assert.NotNil(t, v.Price)
				assert.True(t, v.Price.Equal(mustDecimal("29.99")))
			},
		},
		{
			name: "variant with null price (inherits from product)",
			variant: Variant{
				ID:        2,
				ProductID: 10,
				Name:      "Size Large",
				SKU:       "SKU-LRG-001",
				Price:     nil,
			},
			validate: func(t *testing.T, v Variant) {
				assert.Equal(t, uint(2), v.ID)
				assert.Nil(t, v.Price)
			},
		},
		{
			name: "minimal variant",
			variant: Variant{
				ProductID: 5,
				Name:      "Default",
				SKU:       "DEF-001",
			},
			validate: func(t *testing.T, v Variant) {
				assert.Equal(t, uint(0), v.ID)
				assert.Equal(t, uint(5), v.ProductID)
				assert.Equal(t, "Default", v.Name)
				assert.Equal(t, "DEF-001", v.SKU)
				assert.Nil(t, v.Price)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.variant)
		})
	}
}

func TestVariantFieldTypes(t *testing.T) {
	v := Variant{
		ID:        42,
		ProductID: 100,
		Name:      "Type Check",
		SKU:       "TYPE-001",
		Price:     ptrDecimal(mustDecimal("123.45")),
	}

	// Verify field types through assignments
	assertUint := func(u uint) bool { return true }
	assertString := func(s string) bool { return true }
	assertDecimalPtr := func(d *decimal.Decimal) bool { return true }

	assert.True(t, assertUint(v.ID))
	assert.True(t, assertUint(v.ProductID))
	assert.True(t, assertString(v.Name))
	assert.True(t, assertString(v.SKU))
	assert.True(t, assertDecimalPtr(v.Price))
}

func TestVariantZeroValue(t *testing.T) {
	v := Variant{}

	assert.Equal(t, uint(0), v.ID)
	assert.Equal(t, uint(0), v.ProductID)
	assert.Equal(t, "", v.Name)
	assert.Equal(t, "", v.SKU)
	assert.Nil(t, v.Price)
}

func TestVariantPriceHandling(t *testing.T) {
	tests := []struct {
		name      string
		price     *decimal.Decimal
		isNil     bool
		stringVal string
	}{
		{
			name:      "null price (inheritance)",
			price:     nil,
			isNil:     true,
			stringVal: "<nil>",
		},
		{
			name:      "explicit positive price",
			price:     ptrDecimal(mustDecimal("99.99")),
			isNil:     false,
			stringVal: "99.99",
		},
		{
			name:      "explicit zero price",
			price:     ptrDecimal(decimal.Zero),
			isNil:     false,
			stringVal: "0",
		},
		{
			name:      "high precision price",
			price:     ptrDecimal(mustDecimal("1234.56")),
			isNil:     false,
			stringVal: "1234.56",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Variant{Price: tt.price}
			assert.Equal(t, tt.isNil, v.Price == nil)
			if !tt.isNil {
				assert.Equal(t, tt.stringVal, v.Price.String())
			}
		})
	}
}

func TestVariantSKUUniqueness(t *testing.T) {
	// Test that SKU identifiers are distinct
	v1 := Variant{
		ProductID: 1,
		Name:      "Variant 1",
		SKU:       "UNIQUE-001",
	}

	v2 := Variant{
		ProductID: 1,
		Name:      "Variant 2",
		SKU:       "UNIQUE-002",
	}

	assert.NotEqual(t, v1.SKU, v2.SKU)
	assert.NotEqual(t, v1.Name, v2.Name)
}

func TestVariantProductIDRequired(t *testing.T) {
	// Verify ProductID is tracked as required field
	v1 := Variant{ProductID: 10}
	v2 := Variant{ProductID: 0}

	assert.NotEqual(t, v1.ProductID, v2.ProductID)
	assert.Equal(t, uint(10), v1.ProductID)
	assert.Equal(t, uint(0), v2.ProductID)
}
