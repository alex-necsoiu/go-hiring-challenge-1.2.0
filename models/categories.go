package models

// Category represents a product classification in the catalog.
// Categories group products by type (e.g., Clothing, Shoes, Accessories).
type Category struct {
	// ID is the unique identifier for the category.
	ID uint `gorm:"primaryKey"`
	// Code is a human-readable unique identifier for the category.
	Code string `gorm:"uniqueIndex;not null"`
	// Name is the human-readable display name of the category.
	Name string `gorm:"not null"`
}

// TableName specifies the database table name for the Category model.
func (c *Category) TableName() string {
	return "categories"
}
