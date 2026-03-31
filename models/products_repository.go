package models

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ProductsRepository handles database operations for products.
type ProductsRepository struct {
	db *gorm.DB
}

// NewProductsRepository creates and returns a new ProductsRepository instance.
func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{
		db: db,
	}
}

// GetAllProducts returns all products with their variants from the database.
// Returns a slice of all Product records and an error if the query fails.
func (r *ProductsRepository) GetAllProducts() ([]Product, error) {
	var products []Product
	if err := r.db.Preload("Variants").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// GetProducts returns a paginated list of products with optional filtering.
// It returns the matching products, the total count of all matching records (for pagination),
// and an error if the database query fails.
// Filters applied:
//   - categoryCode: if provided, filters to products in that category
//   - maxPrice: if non-zero, filters to products with price less than maxPrice
// Results are preloaded with Category and Variants relationships.
func (r *ProductsRepository) GetProducts(offset, limit int, categoryCode string, maxPrice decimal.Decimal) ([]Product, int64, error) {
	var products []Product
	var total int64

	query := r.db

	// Apply filters
	if categoryCode != "" {
		query = query.Joins("JOIN categories ON categories.id = products.category_id").
			Where("categories.code = ?", categoryCode)
	}
	if !maxPrice.IsZero() {
		query = query.Where("products.price < ?", maxPrice)
	}

	// Count total matching records
	if err := query.Model(&Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and preloads
	if err := query.Offset(offset).Limit(limit).
		Preload("Category").
		Preload("Variants").
		Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// GetProductByCode retrieves a single product by its code.
// Returns the product with its Category and Variants preloaded, and an error if not found or query fails.
// Returns nil and gorm.ErrRecordNotFound if the product does not exist.
func (r *ProductsRepository) GetProductByCode(code string) (*Product, error) {
	var product Product
	if err := r.db.Where("code = ?", code).
		Preload("Category").
		Preload("Variants").
		First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return &product, nil
}
