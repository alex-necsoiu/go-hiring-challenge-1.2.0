package models

// ProductsRepository handles database operations for products.
type ProductsRepository struct {
	db DBInterface
}

// NewProductsRepository creates and returns a new ProductsRepository instance.
func NewProductsRepository(db DBInterface) *ProductsRepository {
	return &ProductsRepository{
		db: db,
	}
}

// GetProducts returns a paginated list of products with optional filtering.
// It returns the matching products, the total count of all matching records (for pagination),
// and an error if the database query fails.
// Filters applied:
//   - CategoryCode: if provided, filters to products in that category
//   - MaxPrice: if non-zero, filters to products with price less than MaxPrice
// Results are preloaded with Category and Variants relationships.
func (r *ProductsRepository) GetProducts(filter ProductFilter) ([]Product, int64, error) {
	var products []Product
	var total int64

	query := r.db

	// Apply filters
	if filter.CategoryCode != "" {
		query = query.Joins("JOIN categories ON categories.id = products.category_id").
			Where("LOWER(categories.code) = LOWER(?)", filter.CategoryCode)
	}
	if !filter.MaxPrice.IsZero() {
		query = query.Where("products.price < ?", filter.MaxPrice)
	}

	// Count total matching records
	if err := query.Model(&Product{}).Count(&total).GetError(); err != nil {
		return nil, 0, err
	}

	// Apply ordering, pagination and preloads
	if err := query.Order("products.code ASC").Offset(filter.Offset).Limit(filter.Limit).
		Preload("Category").
		Preload("Variants").
		Find(&products).GetError(); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// GetProductByCode retrieves a single product by its code.
// Returns the product with its Category and Variants preloaded, and an error if not found or query fails.
// When the product is not found, the error will be gorm.ErrRecordNotFound.
func (r *ProductsRepository) GetProductByCode(code string) (*Product, error) {
	var product Product
	if err := r.db.Where("code = ?", code).
		Preload("Category").
		Preload("Variants").
		First(&product).GetError(); err != nil {
		return nil, err
	}
	return &product, nil
}
