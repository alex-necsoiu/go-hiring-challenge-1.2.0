package models

import (
	"gorm.io/gorm"
)

// CategoriesRepository handles database operations for categories.
type CategoriesRepository struct {
	db *gorm.DB
}

// NewCategoriesRepository creates and returns a new CategoriesRepository instance.
func NewCategoriesRepository(db *gorm.DB) *CategoriesRepository {
	return &CategoriesRepository{
		db: db,
	}
}

// GetAllCategories returns all categories from the database.
// Returns a slice of all Category records and an error if the query fails.
func (r *CategoriesRepository) GetAllCategories() ([]Category, error) {
	var categories []Category
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// CreateCategory persists a new category to the database.
// Returns an error if the create operation fails (e.g., duplicate code).
func (r *CategoriesRepository) CreateCategory(category *Category) error {
	if err := r.db.Create(category).Error; err != nil {
		return err
	}
	return nil
}
