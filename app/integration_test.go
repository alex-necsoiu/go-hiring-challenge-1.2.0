//go:build integration
// +build integration

package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/mytheresa/go-hiring-challenge/app/catalog"
	"github.com/mytheresa/go-hiring-challenge/app/categories"
	"github.com/mytheresa/go-hiring-challenge/app/database"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ptrDecimal returns a pointer to a decimal.Decimal for use with nullable Variant.Price
func ptrDecimal(d decimal.Decimal) *decimal.Decimal {
	return &d
}

// IntegrationTestSuite manages database setup and teardown for integration tests
type IntegrationTestSuite struct {
	DB                    *gorm.DB
	Close                 func() error
	Router                http.Handler
	ProductsRepository    *models.ProductsRepository
	CategoriesRepository  *models.CategoriesRepository
}

// SetupIntegrationTest initializes database and test fixtures
func SetupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	// Load environment variables
	// Try loading from current directory, then parent directory
	var dotenvErr error
	if err := godotenv.Load(".env"); err != nil {
		// Try parent directory if current fails
		if err := godotenv.Load("../.env"); err != nil {
			dotenvErr = err
		}
	}

	// It's OK if .env doesn't exist - environment variables might be set already
	if dotenvErr != nil {
		t.Logf("Note: .env file not found (env vars may be set): %v", dotenvErr)
	}

	// Initialize database connection
	db, close := database.New(
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_HOST"),
	)

	// Drop existing tables to ensure clean state
	cleanDatabase(t, db)

	// Run migrations to create schema
	runMigrations(t, db)

	// Create repositories
	dbAdapter := models.NewGormDBAdapter(db)
	prodRepo := models.NewProductsRepository(dbAdapter)
	catRepo := models.NewCategoriesRepository(dbAdapter)

	// Create router
	mux := http.NewServeMux()
	catHandler := catalog.NewCatalogHandler(prodRepo)
	catRepoHandler := categories.NewCategoriesHandler(catRepo)

	mux.HandleFunc("GET /catalog", catHandler.HandleGet)
	mux.HandleFunc("GET /catalog/{code}", catHandler.HandleGetByCode)
	mux.HandleFunc("GET /categories", catRepoHandler.HandleGet)
	mux.HandleFunc("POST /categories", catRepoHandler.HandleCreate)

	return &IntegrationTestSuite{
		DB:                   db,
		Close:                close,
		Router:               mux,
		ProductsRepository:   prodRepo,
		CategoriesRepository: catRepo,
	}
}

// TeardownIntegrationTest cleans up database resources
func TeardownIntegrationTest(t *testing.T, suite *IntegrationTestSuite) {
	// Note: table cleanup happens at the START of the next test via cleanDatabase()
	// This is more reliable than cleanup at the end
	if err := suite.Close(); err != nil {
		t.Errorf("Error closing database: %v", err)
	}
}

// runMigrations executes SQL migration files in order, skipping seed data files
func runMigrations(t *testing.T, db *gorm.DB) {
	dir := os.Getenv("POSTGRES_SQL_DIR")
	
	// If env var is set, try it first, but fall back if it doesn't exist
	if dir != "" {
		if _, err := os.Stat(dir); err == nil {
			// Path exists, use it
		} else {
			// Path from env var doesn't exist, try candidates
			dir = ""
		}
	}
	
	// If no valid dir yet, try candidates
	if dir == "" {
		candidates := []string{
			"./sql",
			"../sql",
			"sql",
		}

		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				dir = candidate
				break
			}
		}

		// If still not found, fail with helpful message
		if dir == "" {
			require.Fail(t, "reading migration directory failed", 
				"sql directory not found; tried: ./sql, ../sql, sql, and env var POSTGRES_SQL_DIR")
		}
	}

	files, err := os.ReadDir(dir)
	require.NoError(t, err, "reading migration directory failed")

	// Filter and sort .sql files, but SKIP seed data files for integration tests
	var sqlFiles []os.DirEntry
	skipPatterns := []string{
		"003-product-data.sql",    // Skip seed products
		"005-category-data.sql",   // Skip seed categories
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			// Check if this is a seed data file to skip
			skip := false
			for _, pattern := range skipPatterns {
				if file.Name() == pattern {
					skip = true
					break
				}
			}
			if !skip {
				sqlFiles = append(sqlFiles, file)
			}
		}
	}

	sort.Slice(sqlFiles, func(i, j int) bool {
		return sqlFiles[i].Name() < sqlFiles[j].Name()
	})

	for _, file := range sqlFiles {
		path := filepath.Join(dir, file.Name())
		content, err := os.ReadFile(path)
		require.NoError(t, err, "reading migration file %s failed", file.Name())

		if err := db.Exec(string(content)).Error; err != nil {
			t.Fatalf("executing migration %s failed: %v", file.Name(), err)
		}
	}
}

// truncateTables removes all data from tables while keeping schema
func truncateTables(t *testing.T, db *gorm.DB) {
	// Drop all tables to ensure clean state (in reverse dependency order)
	tables := []string{"product_variants", "products", "categories"}
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
			t.Logf("Warning: dropping table %s failed: %v", table, err)
		}
	}
	
	// Recreate schema by running migrations
	// This is done implicitly by runMigrations on next test run
}

// cleanDatabase drops all tables to ensure a clean state at the start of each test
func cleanDatabase(t *testing.T, db *gorm.DB) {
	tables := []string{"product_variants", "products", "categories"}
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
			// It's OK if tables don't exist yet
			t.Logf("Note: table %s did not exist: %v", table, err)
		}
	}
}

// Helpers for test fixtures

// createTestCategories inserts test categories into the database
func createTestCategories(t *testing.T, db *gorm.DB) []*models.Category {
	categories := []*models.Category{
		{Code: "CLOTHING", Name: "Clothing & Apparel"},
		{Code: "SHOES", Name: "Shoes"},
		{Code: "ACCESSORIES", Name: "Accessories"},
	}

	for _, cat := range categories {
		require.NoError(t, db.Create(cat).Error)
	}

	return categories
}

// createTestProducts inserts test products with variants
func createTestProducts(t *testing.T, db *gorm.DB, categories []*models.Category) []*models.Product {
	products := []*models.Product{
		{
			Code:       "PROD001",
			Name:       "T-Shirt",
			Price:      decimal.NewFromFloat(29.99),
			CategoryID: categories[0].ID, // Clothing
			Variants: []models.Variant{
				{Name: "Small", SKU: "PROD001-S"},
				{Name: "Medium", SKU: "PROD001-M", Price: ptrDecimal(decimal.NewFromFloat(32.99))},
			},
		},
		{
			Code:       "PROD002",
			Name:       "Running Shoes",
			Price:      decimal.NewFromFloat(89.99),
			CategoryID: categories[1].ID, // Shoes
			Variants: []models.Variant{
				{Name: "Size 39", SKU: "PROD002-39"},
				{Name: "Size 40", SKU: "PROD002-40"},
			},
		},
		{
			Code:       "PROD003",
			Name:       "Watch",
			Price:      decimal.NewFromFloat(149.99),
			CategoryID: categories[2].ID, // Accessories
			Variants:   []models.Variant{},
		},
		{
			Code:       "PROD004",
			Name:       "Denim Jacket",
			Price:      decimal.NewFromFloat(79.99),
			CategoryID: categories[0].ID, // Clothing
		},
		{
			Code:       "PROD005",
			Name:       "Sunglasses",
			Price:      decimal.NewFromFloat(159.99),
			CategoryID: categories[2].ID, // Accessories
		},
		{
			Code:       "PROD006",
			Name:       "Boots",
			Price:      decimal.NewFromFloat(119.99),
			CategoryID: categories[1].ID, // Shoes
		},
		{
			Code:       "PROD007",
			Name:       "Sweater",
			Price:      decimal.NewFromFloat(59.99),
			CategoryID: categories[0].ID, // Clothing
		},
		{
			Code:       "PROD008",
			Name:       "Belt",
			Price:      decimal.NewFromFloat(39.99),
			CategoryID: categories[2].ID, // Accessories
		},
	}

	for _, product := range products {
		require.NoError(t, db.Create(product).Error)
	}

	return products
}

// Helper functions for HTTP requests

func doRequest(suite *IntegrationTestSuite, method, path string, body io.Reader) (*http.Response, string) {
	req := httptest.NewRequest(method, path, body)
	w := httptest.NewRecorder()

	suite.Router.ServeHTTP(w, req)

	bodyBytes, _ := io.ReadAll(w.Body)
	return w.Result(), string(bodyBytes)
}

func doJSONRequest(suite *IntegrationTestSuite, method, path string, body interface{}) (*http.Response, string) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(jsonBody)
	}
	return doRequest(suite, method, path, bodyReader)
}

// ============================================================================
// INTEGRATION TESTS
// ============================================================================

// TestIntegration_CatalogEndpoint_ListProducts tests GET /catalog with filtering and pagination
func TestIntegration_CatalogEndpoint_ListProducts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := SetupIntegrationTest(t)
	defer TeardownIntegrationTest(t, suite)

	// Create test data
	categories := createTestCategories(t, suite.DB)
	createTestProducts(t, suite.DB, categories)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:           "list all products with default pagination",
			path:           "/catalog",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]interface{}
				require.NoError(t, json.Unmarshal([]byte(body), &resp))
				data := resp["data"].(map[string]interface{})
				products := data["products"].([]interface{})
				assert.Equal(t, 8, len(products))
				assert.Equal(t, float64(8), data["total"])
			},
		},
		{
			name:           "list products with offset and limit",
			path:           "/catalog?offset=0&limit=3",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]interface{}
				require.NoError(t, json.Unmarshal([]byte(body), &resp))
				data := resp["data"].(map[string]interface{})
				products := data["products"].([]interface{})
				assert.Equal(t, 3, len(products))
				assert.Equal(t, float64(8), data["total"])
			},
		},
		{
			name:           "filter by category",
			path:           "/catalog?category=CLOTHING",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]interface{}
				require.NoError(t, json.Unmarshal([]byte(body), &resp))
				data := resp["data"].(map[string]interface{})
				products := data["products"].([]interface{})
				// PROD001, PROD004, PROD007
				assert.Equal(t, 3, len(products))
				assert.Equal(t, float64(3), data["total"])
			},
		},
		{
			name:           "filter by price less than",
			path:           "/catalog?priceLessThan=80",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]interface{}
				require.NoError(t, json.Unmarshal([]byte(body), &resp))
				data := resp["data"].(map[string]interface{})
				products := data["products"].([]interface{})
				// PROD001 (29.99), PROD004 (79.99), PROD007 (59.99), PROD008 (39.99)
				assert.True(t, len(products) > 0)
				assert.True(t, data["total"].(float64) > 0)
			},
		},
		{
			name:           "combine category and price filters",
			path:           "/catalog?category=SHOES&priceLessThan=120",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]interface{}
				require.NoError(t, json.Unmarshal([]byte(body), &resp))
				data := resp["data"].(map[string]interface{})
				products := data["products"].([]interface{})
				// PROD006 (119.99) - within budget
				assert.True(t, len(products) >= 1)
			},
		},
		{
			name:           "invalid limit - returns bad request",
			path:           "/catalog?limit=abc",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]interface{}
				require.NoError(t, json.Unmarshal([]byte(body), &resp))
				assert.Contains(t, resp["error"].(string), "limit")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := doRequest(suite, http.MethodGet, tt.path, nil)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			tt.checkResponse(t, body)
		})
	}
}

// TestIntegration_CatalogEndpoint_ProductDetails tests GET /catalog/{code}
func TestIntegration_CatalogEndpoint_ProductDetails(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := SetupIntegrationTest(t)
	defer TeardownIntegrationTest(t, suite)

	// Create test data
	categories := createTestCategories(t, suite.DB)
	_ = createTestProducts(t, suite.DB, categories)

	tests := []struct {
		name           string
		code           string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:           "get product with variants",
			code:           "PROD001",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]interface{}
				require.NoError(t, json.Unmarshal([]byte(body), &resp))
				data := resp["data"].(map[string]interface{})
				// Data contains nested product field
				product := data["product"].(map[string]interface{})
				assert.Equal(t, "PROD001", product["code"])
				assert.Equal(t, "T-Shirt", product["name"])
				assert.Equal(t, "CLOTHING", product["category"].(map[string]interface{})["code"])
				variants := product["variants"].([]interface{})
				assert.True(t, len(variants) > 0)
			},
		},
		{
			name:           "get product without variants",
			code:           "PROD003",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]interface{}
				require.NoError(t, json.Unmarshal([]byte(body), &resp))
				data := resp["data"].(map[string]interface{})
				// Data contains nested product field
				product := data["product"].(map[string]interface{})
				assert.Equal(t, "PROD003", product["code"])
				assert.Equal(t, "Watch", product["name"])
			},
		},
		{
			name:           "product not found",
			code:           "INVALID",
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]interface{}
				require.NoError(t, json.Unmarshal([]byte(body), &resp))
				assert.Contains(t, resp["error"].(string), "not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fmt.Sprintf("/catalog/%s", tt.code)
			resp, body := doRequest(suite, http.MethodGet, path, nil)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			tt.checkResponse(t, body)
		})
	}
}

// TestIntegration_CategoriesEndpoint_ListCategories tests GET /categories
func TestIntegration_CategoriesEndpoint_ListCategories(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := SetupIntegrationTest(t)
	defer TeardownIntegrationTest(t, suite)

	// Create test data
	createTestCategories(t, suite.DB)

	resp, body := doRequest(suite, http.MethodGet, "/categories", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(body), &result))

	data := result["data"].(map[string]interface{})
	categories := data["categories"].([]interface{})

	assert.Equal(t, 3, len(categories))
	assert.Equal(t, "CLOTHING", categories[0].(map[string]interface{})["code"])
	assert.Equal(t, "SHOES", categories[1].(map[string]interface{})["code"])
	assert.Equal(t, "ACCESSORIES", categories[2].(map[string]interface{})["code"])
}

// TestIntegration_CategoriesEndpoint_CreateCategory tests POST /categories
func TestIntegration_CategoriesEndpoint_CreateCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := SetupIntegrationTest(t)
	defer TeardownIntegrationTest(t, suite)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name: "create category successfully",
			payload: map[string]interface{}{
				"code": "ELECTRONICS",
				"name": "Electronics & Gadgets",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]interface{}
				require.NoError(t, json.Unmarshal([]byte(body), &resp))
				assert.Contains(t, resp, "data")
			},
		},
		{
			name: "create category with duplicate code",
			payload: map[string]interface{}{
				"code": "ELECTRONICS",
				"name": "Different Name",
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]interface{}
				require.NoError(t, json.Unmarshal([]byte(body), &resp))
				assert.Contains(t, resp["error"].(string), "already exists")
			},
		},
		{
			name: "create category with missing code",
			payload: map[string]interface{}{
				"name": "Missing Code",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]interface{}
				require.NoError(t, json.Unmarshal([]byte(body), &resp))
				assert.Contains(t, resp["error"].(string), "code")
			},
		},
		{
			name: "create category with missing name",
			payload: map[string]interface{}{
				"code": "NO-NAME",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]interface{}
				require.NoError(t, json.Unmarshal([]byte(body), &resp))
				assert.Contains(t, resp["error"].(string), "name")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := doJSONRequest(suite, http.MethodPost, "/categories", tt.payload)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			tt.checkResponse(t, body)
		})
	}
}

// TestIntegration_FullWorkflow tests complete workflows
func TestIntegration_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := SetupIntegrationTest(t)
	defer TeardownIntegrationTest(t, suite)

	// 1. Create categories
	resp, _ := doJSONRequest(suite, http.MethodPost, "/categories", map[string]interface{}{
		"code": "SPORTS",
		"name": "Sports & Outdoors",
	})
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// 2. List categories (should have 1)
	resp, body := doRequest(suite, http.MethodGet, "/categories", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var result map[string]interface{}
	json.Unmarshal([]byte(body), &result)
	categories := result["data"].(map[string]interface{})["categories"].([]interface{})
	assert.Equal(t, 1, len(categories))

	// 3. List catalog (empty initially)
	resp, body = doRequest(suite, http.MethodGet, "/catalog", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	json.Unmarshal([]byte(body), &result)
	data := result["data"].(map[string]interface{})
	products := data["products"].([]interface{})
	assert.Equal(t, 0, len(products))
	assert.Equal(t, float64(0), data["total"])
}
