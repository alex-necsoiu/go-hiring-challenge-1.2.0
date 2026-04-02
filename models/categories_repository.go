package models

// CategoriesRepository handles database operations for categories.
type CategoriesRepository struct {
	db DBInterface
}

// NewCategoriesRepository creates and returns a new CategoriesRepository instance.
func NewCategoriesRepository(db DBInterface) *CategoriesRepository {
	return &CategoriesRepository{
		db: db,
	}
}

// GetAllCategories returns all categories from the database in code order.
// Results are ordered by code (ascending) for deterministic API responses.
// Returns a slice of all Category records and an error if the query fails.
func (r *CategoriesRepository) GetAllCategories() ([]Category, error) {
	var categories []Category
	if err := r.db.Order("code ASC").Find(&categories).GetError(); err != nil {
		return nil, err
	}
	return categories, nil
}

// CreateCategory persists a new category to the database.
// The category pointer is populated with the generated ID after successful insert.
// Returns an error if the create operation fails. On constraint violations (e.g., duplicate code),
// the error message will contain "UNIQUE constraint failed" or similar database-specific text.
// Callers should parse the error message to distinguish constraint violations from other errors.
func (r *CategoriesRepository) CreateCategory(category *Category) error {
	if err := r.db.Create(category).GetError(); err != nil {
		return err
	}
	return nil
}
