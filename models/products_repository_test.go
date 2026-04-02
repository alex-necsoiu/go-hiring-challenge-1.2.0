package models

import (
	"fmt"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// MockDB is a mock implementation of the database for testing repositories
type MockDB struct {
	products   []Product
	categories []Category
	shouldFail bool
	failError  error
}

func NewMockDB(products []Product, categories []Category) *MockDB {
	return &MockDB{
		products:   products,
		categories: categories,
	}
}

func (m *MockDB) WithError(err error) *MockDB {
	m.failError = err
	m.shouldFail = true
	return m
}

func TestNewProductsRepository(t *testing.T) {
	mockDB := &productsMockGormDB{
		MockDB: &MockDB{},
	}
	repo := NewProductsRepository(mockDB)

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.db)
}

func TestProductsRepositoryGetProducts(t *testing.T) {
	testProducts := []Product{
		{
			ID:         1,
			Code:       "CLOTHING001",
			Name:       "Shirt",
			Price:      mustDecimal("25.00"),
			CategoryID: 1,
			Category:   Category{ID: 1, Code: "clothing", Name: "Clothing"},
		},
		{
			ID:         2,
			Code:       "SHOES001",
			Name:       "Sneakers",
			Price:      mustDecimal("75.00"),
			CategoryID: 2,
			Category:   Category{ID: 2, Code: "shoes", Name: "Shoes"},
		},
		{
			ID:         3,
			Code:       "CLOTHING002",
			Name:       "Pants",
			Price:      mustDecimal("50.00"),
			CategoryID: 1,
			Category:   Category{ID: 1, Code: "clothing", Name: "Clothing"},
		},
	}

	tests := []struct {
		name     string
		filter   ProductFilter
		validate func(t *testing.T, products []Product, total int64, err error)
	}{
		{
			name: "returns all products with total count",
			filter: ProductFilter{
				Offset: 0,
				Limit:  10,
			},
			validate: func(t *testing.T, products []Product, total int64, err error) {
				assert.NoError(t, err)
				assert.Len(t, products, 3)
				assert.Equal(t, int64(3), total)
			},
		},
		{
			name: "applies pagination offset and limit",
			filter: ProductFilter{
				Offset: 1,
				Limit:  1,
			},
			validate: func(t *testing.T, products []Product, total int64, err error) {
				assert.NoError(t, err)
				assert.Len(t, products, 1)
				assert.Equal(t, int64(3), total)
			},
		},
		{
			name: "filters by category code",
			filter: ProductFilter{
				Offset:       0,
				Limit:        10,
				CategoryCode: "clothing",
			},
			validate: func(t *testing.T, products []Product, total int64, err error) {
				assert.NoError(t, err)
				assert.Equal(t, int64(2), total)
			},
		},
		{
			name: "filters by max price",
			filter: ProductFilter{
				Offset:   0,
				Limit:    10,
				MaxPrice: mustDecimal("30.00"),
			},
			validate: func(t *testing.T, products []Product, total int64, err error) {
				assert.NoError(t, err)
				assert.Equal(t, int64(1), total)
			},
		},
		{
			name: "combines category and price filters",
			filter: ProductFilter{
				Offset:       0,
				Limit:        10,
				CategoryCode: "clothing",
				MaxPrice:     mustDecimal("40.00"),
			},
			validate: func(t *testing.T, products []Product, total int64, err error) {
				assert.NoError(t, err)
				// Only CLOTHING001 (25.00) matches both clothing category AND price < 40
				assert.Equal(t, int64(1), total)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDB(testProducts, []Category{})
			repo := &ProductsRepository{
				db: &productsMockGormDB{MockDB: mockDB},
			}

			products, total, err := repo.GetProducts(tt.filter)
			tt.validate(t, products, total, err)
		})
	}
}

func TestProductsRepositoryGetProductByCode(t *testing.T) {
	testProducts := []Product{
		{
			ID:         1,
			Code:       "PROD001",
			Name:       "Product One",
			Price:      mustDecimal("99.99"),
			CategoryID: 1,
			Category:   Category{ID: 1, Code: "test", Name: "Test"},
			Variants: []Variant{
				{ID: 1, ProductID: 1, Name: "Variant 1", SKU: "SKU001"},
			},
		},
		{
			ID:   2,
			Code: "PROD002",
			Name: "Product Two",
		},
	}

	tests := []struct {
		name      string
		code      string
		validate  func(t *testing.T, product *Product, err error)
		expectErr bool
	}{
		{
			name: "returns product by code with relationships",
			code: "PROD001",
			validate: func(t *testing.T, product *Product, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, product)
				assert.Equal(t, "PROD001", product.Code)
				assert.Equal(t, "Product One", product.Name)
				assert.Len(t, product.Variants, 1)
			},
		},
		{
			name: "returns product without relationships",
			code: "PROD002",
			validate: func(t *testing.T, product *Product, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, product)
				assert.Equal(t, "PROD002", product.Code)
			},
		},
		{
			name:      "returns error for nonexistent product",
			code:      "NONEXISTENT",
			expectErr: true,
			validate: func(t *testing.T, product *Product, err error) {
				assert.Error(t, err)
				assert.Nil(t, product)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDB(testProducts, []Category{})
			repo := &ProductsRepository{
				db: &productsMockGormDB{MockDB: mockDB},
			}

			product, err := repo.GetProductByCode(tt.code)
			tt.validate(t, product, err)
		})
	}
}

// TestProductsRepositoryEdgeCases covers additional edge cases for coverage
func TestProductsRepositoryEdgeCases(t *testing.T) {
	t.Run("GetProducts with large offset", func(t *testing.T) {
		testProducts := []Product{
			{ID: 1, Code: "P1", Price: mustDecimal("100.00")},
		}
		mockDB := NewMockDB(testProducts, []Category{})
		repo := &ProductsRepository{
			db: &productsMockGormDB{MockDB: mockDB},
		}

		products, total, err := repo.GetProducts(ProductFilter{
			Offset: 100,
			Limit:  10,
		})
		assert.NoError(t, err)
		assert.Len(t, products, 0)
		assert.Equal(t, int64(1), total)
	})

	t.Run("GetProducts with zero limit", func(t *testing.T) {
		testProducts := []Product{
			{ID: 1, Code: "P1", Price: mustDecimal("100.00")},
			{ID: 2, Code: "P2", Price: mustDecimal("200.00")},
		}
		mockDB := NewMockDB(testProducts, []Category{})
		repo := &ProductsRepository{
			db: &productsMockGormDB{MockDB: mockDB},
		}

		products, total, err := repo.GetProducts(ProductFilter{
			Offset: 0,
			Limit:  0,
		})
		assert.NoError(t, err)
		assert.Len(t, products, 2)
		assert.Equal(t, int64(2), total)
	})

	t.Run("GetProductByCode returns correct product with variants", func(t *testing.T) {
		testProducts := []Product{
			{
				ID:   1,
				Code: "PROD001",
				Name: "Test Product",
				Price: mustDecimal("50.00"),
				Category: Category{ID: 1, Code: "test", Name: "Test"},
				Variants: []Variant{
					{ID: 1, ProductID: 1, Name: "V1", SKU: "SKU1"},
					{ID: 2, ProductID: 1, Name: "V2", SKU: "SKU2"},
				},
			},
		}
		mockDB := NewMockDB(testProducts, []Category{})
		repo := &ProductsRepository{
			db: &productsMockGormDB{MockDB: mockDB},
		}

		product, err := repo.GetProductByCode("PROD001")
		assert.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, "PROD001", product.Code)
		assert.Len(t, product.Variants, 2)
	})

	t.Run("GetProducts category filter only excludes other categories", func(t *testing.T) {
		testProducts := []Product{
			{ID: 1, Code: "CLOTHING001", CategoryID: 1, Price: mustDecimal("25.00"), Category: Category{ID: 1, Code: "clothing"}},
			{ID: 2, Code: "CLOTHING002", CategoryID: 1, Price: mustDecimal("35.00"), Category: Category{ID: 1, Code: "clothing"}},
			{ID: 3, Code: "SHOES001", CategoryID: 2, Price: mustDecimal("75.00"), Category: Category{ID: 2, Code: "shoes"}},
		}
		mockDB := NewMockDB(testProducts, []Category{})
		repo := &ProductsRepository{
			db: &productsMockGormDB{MockDB: mockDB},
		}

		products, total, err := repo.GetProducts(ProductFilter{
			Offset:       0,
			Limit:        10,
			CategoryCode: "clothing",
		})
		assert.NoError(t, err)
		assert.Len(t, products, 2)
		assert.Equal(t, int64(2), total)
		assert.True(t, len(products) > 0 && products[0].Category.Code == "clothing")
	})

	t.Run("GetProducts handles combined category and price filtering", func(t *testing.T) {
		testProducts := []Product{
			{ID: 1, Code: "C1", CategoryID: 1, Price: mustDecimal("10.00"), Category: Category{ID: 1, Code: "cat1"}},
			{ID: 2, Code: "C2", CategoryID: 1, Price: mustDecimal("50.00"), Category: Category{ID: 1, Code: "cat1"}},
			{ID: 3, Code: "C3", CategoryID: 1, Price: mustDecimal("100.00"), Category: Category{ID: 1, Code: "cat1"}},
		}
		mockDB := NewMockDB(testProducts, []Category{})
		repo := &ProductsRepository{
			db: &productsMockGormDB{MockDB: mockDB},
		}

		products, total, err := repo.GetProducts(ProductFilter{
			Offset:       0,
			Limit:        10,
			CategoryCode: "cat1",
			MaxPrice:     mustDecimal("75.00"),
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, products, 2)
	})
}

// productsMockGormDB implements minimal GORM-like interface for products testing
type productsMockGormDB struct {
	*MockDB
	lastError        error
	offset           int
	limit            int
	categoryFilter   string
	priceFilter      string
	codeFilter       string
	hasJoin          bool
	hasWhere         bool
	hasCodeWhere     bool
	orderBy          string
}

func (m *productsMockGormDB) filterProducts() []Product {
	results := m.products

	// Apply category filter (case-insensitive)
	if m.hasJoin && m.categoryFilter != "" {
		filtered := []Product{}
		for _, p := range results {
			// First check if product has Category set with matching code (case-insensitive)
			if p.Category.Code != "" && equalsIgnoreCase(p.Category.Code, m.categoryFilter) {
				filtered = append(filtered, p)
				continue
			}

			// Otherwise, find category by ID in MockDB.categories (case-insensitive)
			for _, cat := range m.categories {
				if cat.ID == p.CategoryID && equalsIgnoreCase(cat.Code, m.categoryFilter) {
					filtered = append(filtered, p)
					break
				}
			}
		}
		results = filtered
	}

	// Apply price filter
	if m.hasWhere && m.priceFilter != "" {
		filtered := []Product{}
		for _, p := range results {
			// Simple price comparison (less than maxPrice)
			if p.Price.LessThan(decimal.RequireFromString(m.priceFilter)) {
				filtered = append(filtered, p)
			}
		}
		results = filtered
	}

	// Apply ordering
	if m.orderBy != "" {
		if m.orderBy == "products.code ASC" {
			// Sort by code in ascending order
			for i := 0; i < len(results); i++ {
				for j := i + 1; j < len(results); j++ {
					if results[j].Code < results[i].Code {
						results[i], results[j] = results[j], results[i]
					}
				}
			}
		}
	}

	return results
}

// equalsIgnoreCase performs case-insensitive string comparison using strings.EqualFold.
// This correctly handles Unicode characters and follows Go standard library practices.
func equalsIgnoreCase(a, b string) bool {
	return strings.EqualFold(a, b)
}

func (m *productsMockGormDB) Preload(column string, conditions ...interface{}) DBInterface {
	return m
}

func (m *productsMockGormDB) Find(dest interface{}) DBInterface {
	if m.shouldFail {
		m.lastError = m.failError
		return m
	}
	if products, ok := dest.(*[]Product); ok {
		allProducts := m.filterProducts()

		// Apply offset and limit
		start := m.offset
		end := start + m.limit
		if m.limit == 0 {
			end = len(allProducts)
		}
		if start >= len(allProducts) {
			*products = []Product{}
		} else if end > len(allProducts) {
			*products = allProducts[start:]
		} else {
			*products = allProducts[start:end]
		}
	}
	m.lastError = nil
	return m
}

func (m *productsMockGormDB) Joins(query string, args ...interface{}) DBInterface {
	newMock := &productsMockGormDB{
		MockDB:         m.MockDB,
		offset:         m.offset,
		limit:          m.limit,
		categoryFilter: m.categoryFilter,
		priceFilter:    m.priceFilter,
		codeFilter:     m.codeFilter,
		hasJoin:        true,
		hasWhere:       m.hasWhere,
		hasCodeWhere:   m.hasCodeWhere,
		orderBy:        m.orderBy,
	}
	return newMock
}

func (m *productsMockGormDB) Where(query interface{}, args ...interface{}) DBInterface {
	newMock := &productsMockGormDB{
		MockDB:         m.MockDB,
		offset:         m.offset,
		limit:          m.limit,
		categoryFilter: m.categoryFilter,
		priceFilter:    m.priceFilter,
		codeFilter:     m.codeFilter,
		hasJoin:        m.hasJoin,
		hasWhere:       true,
		hasCodeWhere:   m.hasCodeWhere,
		orderBy:        m.orderBy,
	}

	// Extract filter values from args
	if len(args) > 0 {
		if queryStr, ok := query.(string); ok {
			// Handle both case-sensitive and case-insensitive category filters
			if queryStr == "categories.code = ?" || queryStr == "LOWER(categories.code) = LOWER(?)" {
				if code, ok := args[0].(string); ok {
					newMock.categoryFilter = code
				}
			} else if queryStr == "products.price < ?" {
				if val, ok := args[0].(decimal.Decimal); ok {
					newMock.priceFilter = val.String()
				}
			} else if queryStr == "code = ?" {
				if code, ok := args[0].(string); ok {
					newMock.codeFilter = code
					newMock.hasCodeWhere = true
				}
			}
		}
	}

	return newMock
}

func (m *productsMockGormDB) Model(value interface{}) DBInterface {
	return m
}

func (m *productsMockGormDB) Count(count *int64) DBInterface {
	filtered := m.filterProducts()
	*count = int64(len(filtered))
	m.lastError = nil
	return m
}

func (m *productsMockGormDB) Offset(offset int) DBInterface {
	newMock := &productsMockGormDB{
		MockDB:         m.MockDB,
		offset:         offset,
		limit:          m.limit,
		categoryFilter: m.categoryFilter,
		priceFilter:    m.priceFilter,
		codeFilter:     m.codeFilter,
		hasJoin:        m.hasJoin,
		hasWhere:       m.hasWhere,
		hasCodeWhere:   m.hasCodeWhere,
		orderBy:        m.orderBy,
	}
	return newMock
}

func (m *productsMockGormDB) Order(value interface{}) DBInterface {
	newMock := &productsMockGormDB{
		MockDB:         m.MockDB,
		offset:         m.offset,
		limit:          m.limit,
		categoryFilter: m.categoryFilter,
		priceFilter:    m.priceFilter,
		codeFilter:     m.codeFilter,
		hasJoin:        m.hasJoin,
		hasWhere:       m.hasWhere,
		hasCodeWhere:   m.hasCodeWhere,
		orderBy:        fmt.Sprintf("%v", value),
	}
	return newMock
}

func (m *productsMockGormDB) Limit(limit int) DBInterface {
	newMock := &productsMockGormDB{
		MockDB:         m.MockDB,
		offset:         m.offset,
		limit:          limit,
		categoryFilter: m.categoryFilter,
		priceFilter:    m.priceFilter,
		codeFilter:     m.codeFilter,
		hasJoin:        m.hasJoin,
		hasWhere:       m.hasWhere,
		hasCodeWhere:   m.hasCodeWhere,
		orderBy:        m.orderBy,
	}
	return newMock
}

func (m *productsMockGormDB) First(dest interface{}) DBInterface {
	if product, ok := dest.(*Product); ok {
		// If filtering by code
		if m.hasCodeWhere && m.codeFilter != "" {
			for _, p := range m.products {
				if p.Code == m.codeFilter {
					*product = p
					m.lastError = nil
					return m
				}
			}
			// Not found
			m.lastError = gorm.ErrRecordNotFound
			return m
		}

		// Default: return first product
		if len(m.products) == 0 {
			m.lastError = gorm.ErrRecordNotFound
			return m
		}
		*product = m.products[0]
		m.lastError = nil
		return m
	}

	if len(m.products) == 0 {
		m.lastError = gorm.ErrRecordNotFound
		return m
	}
	m.lastError = nil
	return m
}

func (m *productsMockGormDB) Create(value interface{}) DBInterface {
	if m.shouldFail {
		m.lastError = m.failError
		return m
	}
	m.lastError = nil
	return m
}

func (m *productsMockGormDB) GetError() error {
	return m.lastError
}
