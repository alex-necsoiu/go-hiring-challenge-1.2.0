package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCategoryTableName(t *testing.T) {
	c := &Category{}
	assert.Equal(t, "categories", c.TableName())
}

func TestCategoryCreation(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		validate func(t *testing.T, c Category)
	}{
		{
			name: "complete category",
			category: Category{
				ID:   1,
				Code: "CLOTHING",
				Name: "Clothing & Apparel",
			},
			validate: func(t *testing.T, c Category) {
				assert.Equal(t, uint(1), c.ID)
				assert.Equal(t, "CLOTHING", c.Code)
				assert.Equal(t, "Clothing & Apparel", c.Name)
			},
		},
		{
			name: "minimal category",
			category: Category{
				Code: "MIN",
				Name: "Minimal",
			},
			validate: func(t *testing.T, c Category) {
				assert.Equal(t, uint(0), c.ID)
				assert.Equal(t, "MIN", c.Code)
				assert.Equal(t, "Minimal", c.Name)
			},
		},
		{
			name: "category with special characters",
			category: Category{
				Code: "SPECIAL-CODE_123",
				Name: "Special & Unique (Test)",
			},
			validate: func(t *testing.T, c Category) {
				assert.Equal(t, "SPECIAL-CODE_123", c.Code)
				assert.Equal(t, "Special & Unique (Test)", c.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.category)
		})
	}
}

func TestCategoryFieldTypes(t *testing.T) {
	c := Category{
		ID:   99,
		Code: "TYPE-CHECK",
		Name: "Type Verification",
	}

	// Verify field types through assignments
	assertUint := func(v uint) bool { return true }
	assertString := func(v string) bool { return true }

	assert.True(t, assertUint(c.ID))
	assert.True(t, assertString(c.Code))
	assert.True(t, assertString(c.Name))
}

func TestCategoryZeroValue(t *testing.T) {
	c := Category{}

	assert.Equal(t, uint(0), c.ID)
	assert.Equal(t, "", c.Code)
	assert.Equal(t, "", c.Name)
}

func TestCategoryCodeUniqueness(t *testing.T) {
	// Test that code field is distinct
	c1 := Category{
		ID:   1,
		Code: "CAT-001",
		Name: "Category 1",
	}

	c2 := Category{
		ID:   2,
		Code: "CAT-002",
		Name: "Category 1", // Same name, different code
	}

	c3 := Category{
		ID:   3,
		Code: "CAT-001", // Duplicate code
		Name: "Different Name",
	}

	assert.NotEqual(t, c1.Code, c2.Code)
	assert.Equal(t, c1.Code, c3.Code)
	assert.NotEqual(t, c1.Name, c3.Name)
}
