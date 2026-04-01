package models

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCategoriesRepository(t *testing.T) {
	mockDB := &categoriesMockGormDB{
		MockDB: &MockDB{},
	}
	repo := NewCategoriesRepository(mockDB)

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.db)
}

func TestCategoriesRepositoryGetAllCategories(t *testing.T) {
	tests := []struct {
		name       string
		categories []Category
		setup      func(*MockDB)
		expectErr  bool
		validate   func(t *testing.T, categories []Category)
	}{
		{
			name: "returns all categories",
			categories: []Category{
				{ID: 1, Code: "CLOTHING", Name: "Clothing & Apparel"},
				{ID: 2, Code: "SHOES", Name: "Shoes"},
				{ID: 3, Code: "ACCESSORIES", Name: "Accessories"},
			},
			setup:     func(m *MockDB) {},
			expectErr: false,
			validate: func(t *testing.T, categories []Category) {
				assert.Len(t, categories, 3)
				assert.Equal(t, "CLOTHING", categories[0].Code)
				assert.Equal(t, "Shoes", categories[1].Name)
			},
		},
		{
			name:      "returns empty slice when no categories",
			categories: []Category{},
			setup:     func(m *MockDB) {},
			validate: func(t *testing.T, categories []Category) {
				assert.Len(t, categories, 0)
			},
		},
		{
			name: "handles database error",
			setup: func(m *MockDB) {
				m.shouldFail = true
				m.failError = errors.New("database error")
			},
			expectErr: true,
			validate: func(t *testing.T, categories []Category) {
				assert.Nil(t, categories)
			},
		},
		{
			name: "returns single category",
			categories: []Category{
				{ID: 1, Code: "TEST", Name: "Test Category"},
			},
			setup:     func(m *MockDB) {},
			expectErr: false,
			validate: func(t *testing.T, categories []Category) {
				assert.Len(t, categories, 1)
				assert.Equal(t, uint(1), categories[0].ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDB([]Product{}, tt.categories)
			tt.setup(mockDB)

			repo := &CategoriesRepository{
				db: &categoriesMockGormDB{MockDB: mockDB},
			}

			categories, err := repo.GetAllCategories()

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.validate(t, categories)
		})
	}
}

func TestCategoriesRepositoryCreateCategory(t *testing.T) {
	tests := []struct {
		name       string
		category   *Category
		setup      func(*MockDB)
		expectErr  bool
		validate   func(t *testing.T, err error)
	}{
		{
			name: "creates category successfully",
			category: &Category{
				Code: "NEW-CAT",
				Name: "New Category",
			},
			setup:     func(m *MockDB) {},
			expectErr: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "creates category with ID",
			category: &Category{
				ID:   99,
				Code: "WITH-ID",
				Name: "With ID",
			},
			setup:     func(m *MockDB) {},
			expectErr: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "handles database error (e.g., duplicate code)",
			category: &Category{
				Code: "DUPLICATE",
				Name: "Duplicate Category",
			},
			setup: func(m *MockDB) {
				m.shouldFail = true
				m.failError = errors.New("UNIQUE constraint failed: categories.code")
			},
			expectErr: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "UNIQUE constraint")
			},
		},
		{
			name: "creates another category after first",
			category: &Category{
				Code: "FOLLOW-UP",
				Name: "Follow-Up Category",
			},
			setup:     func(m *MockDB) {},
			expectErr: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDB([]Product{}, []Category{})
			tt.setup(mockDB)

			repo := &CategoriesRepository{
				db: &categoriesMockGormDB{MockDB: mockDB},
			}

			err := repo.CreateCategory(tt.category)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.validate(t, err)
		})
	}
}

func TestCategoriesRepositoryIntegration(t *testing.T) {
	// Test sequence: create categories, then retrieve them
	t.Run("create then retrieve categories", func(t *testing.T) {
		mockDB := NewMockDB([]Product{}, []Category{})

		repo := &CategoriesRepository{
			db: &categoriesMockGormDB{MockDB: mockDB},
		}

		// Create new categories
		cat1 := &Category{Code: "COLOR", Name: "Colors"}
		cat2 := &Category{Code: "SIZE", Name: "Sizes"}

		err1 := repo.CreateCategory(cat1)
		err2 := repo.CreateCategory(cat2)

		assert.NoError(t, err1)
		assert.NoError(t, err2)

		// Retrieve all categories
		categories, err := repo.GetAllCategories()
		assert.NoError(t, err)
		assert.NotNil(t, categories)
	})
}

// categoriesMockGormDB implements minimal GORM-like interface for categories testing
type categoriesMockGormDB struct {
	*MockDB
	lastError error
}

func (m *categoriesMockGormDB) Preload(column string, conditions ...interface{}) DBInterface {
	return m
}

func (m *categoriesMockGormDB) Find(dest interface{}) DBInterface {
	if m.shouldFail {
		m.lastError = m.failError
		return m
	}
	if categories, ok := dest.(*[]Category); ok {
		*categories = m.categories
	}
	m.lastError = nil
	return m
}

func (m *categoriesMockGormDB) Joins(query string, args ...interface{}) DBInterface {
	return m
}

func (m *categoriesMockGormDB) Where(query interface{}, args ...interface{}) DBInterface {
	return m
}

func (m *categoriesMockGormDB) Model(value interface{}) DBInterface {
	return m
}

func (m *categoriesMockGormDB) Count(count *int64) DBInterface {
	*count = int64(len(m.categories))
	return m
}

func (m *categoriesMockGormDB) Offset(offset int) DBInterface {
	return m
}

func (m *categoriesMockGormDB) Limit(limit int) DBInterface {
	return m
}

func (m *categoriesMockGormDB) First(dest interface{}) DBInterface {
	return m
}

func (m *categoriesMockGormDB) Create(value interface{}) DBInterface {
	if m.shouldFail {
		m.lastError = m.failError
		return m
	}
	if category, ok := value.(*Category); ok {
		m.categories = append(m.categories, *category)
	}
	m.lastError = nil
	return m
}

func (m *categoriesMockGormDB) Order(value interface{}) DBInterface {
	return m
}

func (m *categoriesMockGormDB) GetError() error {
	return m.lastError
}
