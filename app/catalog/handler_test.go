package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/mytheresa/go-hiring-challenge/models"
	"gorm.io/gorm"
)

// mockProductRepository is a test double implementing ProductRepository interface
type mockProductRepository struct {
	getProductsFunc      func(filter models.ProductFilter) ([]models.Product, int64, error)
	getProductByCodeFunc func(code string) (*models.Product, error)
}

func (m *mockProductRepository) GetProducts(filter models.ProductFilter) ([]models.Product, int64, error) {
	return m.getProductsFunc(filter)
}

func (m *mockProductRepository) GetProductByCode(code string) (*models.Product, error) {
	return m.getProductByCodeFunc(code)
}

func TestHandleGet(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockFunc       func() ([]models.Product, int64, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
		checkFilter    func(t *testing.T, filter models.ProductFilter)
	}{
		{
			name:        "happy path - returns 200 with products",
			queryParams: "?offset=0&limit=10",
			mockFunc: func() ([]models.Product, int64, error) {
				return []models.Product{
					{
						ID:   1,
						Code: "PROD001",
						Price: decimal.NewFromFloat(10.99),
						CategoryID: 1,
						Category: models.Category{
							ID:   1,
							Code: "clothing",
							Name: "Clothing",
						},
						Variants: []models.Variant{},
					},
				}, 1, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp CatalogResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, 1, len(resp.Data.Products))
				assert.Equal(t, "PROD001", resp.Data.Products[0].Code)
				assert.Equal(t, "10.99", resp.Data.Products[0].Price)
				assert.Equal(t, int64(1), resp.Data.Total)
			},
			checkFilter: func(t *testing.T, filter models.ProductFilter) {
				assert.Equal(t, 0, filter.Offset)
				assert.Equal(t, 10, filter.Limit)
			},
		},
		{
			name:        "empty list - returns 200",
			queryParams: "?offset=0&limit=10",
			mockFunc: func() ([]models.Product, int64, error) {
				return []models.Product{}, 0, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp CatalogResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, 0, len(resp.Data.Products))
				assert.Equal(t, int64(0), resp.Data.Total)
			},
			checkFilter: nil,
		},
		{
			name:           "invalid limit (not numeric) - returns 400",
			queryParams:    "?offset=0&limit=abc",
			mockFunc:       func() ([]models.Product, int64, error) { return nil, 0, nil },
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp["error"], "limit")
			},
			checkFilter: nil,
		},
		{
			name:           "invalid offset (not numeric) - returns 400",
			queryParams:    "?offset=xyz&limit=10",
			mockFunc:       func() ([]models.Product, int64, error) { return nil, 0, nil },
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp["error"], "offset")
			},
			checkFilter: nil,
		},
		{
			name:           "limit > 100 - returns 400",
			queryParams:    "?offset=0&limit=101",
			mockFunc:       func() ([]models.Product, int64, error) { return nil, 0, nil },
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp["error"], "limit")
			},
			checkFilter: nil,
		},
		{
			name:           "limit < 1 - returns 400",
			queryParams:    "?offset=0&limit=0",
			mockFunc:       func() ([]models.Product, int64, error) { return nil, 0, nil },
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp["error"], "limit")
			},
			checkFilter: nil,
		},
		{
			name:        "correct total and pagination fields",
			queryParams: "?offset=5&limit=20",
			mockFunc: func() ([]models.Product, int64, error) {
				return []models.Product{}, 100, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp CatalogResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, int64(100), resp.Data.Total)
				assert.Equal(t, 5, resp.Data.Offset)
				assert.Equal(t, 20, resp.Data.Limit)
			},
			checkFilter: nil,
		},
		{
			name:        "repository error - returns 500",
			queryParams: "?offset=0&limit=10",
			mockFunc: func() ([]models.Product, int64, error) {
				return nil, 0, assert.AnError
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp["error"], "")
			},
			checkFilter: nil,
		},
		{
			name:        "category filter passed to repository",
			queryParams: "?offset=0&limit=10&category=clothing",
			mockFunc: func() ([]models.Product, int64, error) {
				return []models.Product{}, 0, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp CatalogResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
			},
			checkFilter: func(t *testing.T, filter models.ProductFilter) {
				assert.Equal(t, "clothing", filter.CategoryCode)
			},
		},
		{
			name:        "priceLessThan filter passed to repository",
			queryParams: "?offset=0&limit=10&priceLessThan=15.00",
			mockFunc: func() ([]models.Product, int64, error) {
				return []models.Product{}, 0, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp CatalogResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
			},
			checkFilter: func(t *testing.T, filter models.ProductFilter) {
				expected := decimal.NewFromFloat(15.00)
				assert.True(t, filter.MaxPrice.Equal(expected))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedFilter := models.ProductFilter{}
			mock := &mockProductRepository{
				getProductsFunc: func(filter models.ProductFilter) ([]models.Product, int64, error) {
					capturedFilter = filter
					return tt.mockFunc()
				},
			}

			handler := &CatalogHandler{repo: mock}
			req := httptest.NewRequest("GET", "/catalog"+tt.queryParams, nil)
			recorder := httptest.NewRecorder()

			handler.HandleGet(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			tt.checkResponse(t, recorder.Body.String())
			if tt.checkFilter != nil {
				tt.checkFilter(t, capturedFilter)
			}
		})
	}
}

func TestHandleGetByCode(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		mockFunc       func(code string) (*models.Product, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name: "happy path - returns 200 with product detail",
			code: "PROD001",
			mockFunc: func(code string) (*models.Product, error) {
				return &models.Product{
					ID:   1,
					Code: "PROD001",
					Price: decimal.NewFromFloat(10.99),
					CategoryID: 1,
					Category: models.Category{
						ID:   1,
						Code: "clothing",
						Name: "Clothing",
					},
					Variants: []models.Variant{
						{
							ID: 1,
							Name: "Variant A",
							SKU: "SKU001A",
							Price: decimal.NewFromFloat(11.99),
						},
						{
							ID: 2,
							Name: "Variant B",
							SKU: "SKU001B",
							Price: decimal.Zero,
						},
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp ProductDetailResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "PROD001", resp.Data.Code)
				assert.Equal(t, "10.99", resp.Data.Price)
				assert.Equal(t, "clothing", resp.Data.Category.Code)
				assert.Equal(t, 2, len(resp.Data.Variants))
				// Variant with price should have that price
				assert.Equal(t, "11.99", resp.Data.Variants[0].Price)
				// Variant without price should inherit product price
				assert.Equal(t, "10.99", resp.Data.Variants[1].Price)
			},
		},
		{
			name: "product not found - returns 404",
			code: "NONEXISTENT",
			mockFunc: func(code string) (*models.Product, error) {
				return nil, gorm.ErrRecordNotFound
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp["error"], "not found")
			},
		},
		{
			name: "repository error - returns 500",
			code: "PROD001",
			mockFunc: func(code string) (*models.Product, error) {
				return nil, assert.AnError
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockProductRepository{
				getProductByCodeFunc: tt.mockFunc,
			}

			handler := &CatalogHandler{repo: mock}
			req := httptest.NewRequest("GET", "/catalog/"+tt.code, nil)
			req.SetPathValue("code", tt.code)
			recorder := httptest.NewRecorder()

			handler.HandleGetByCode(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			tt.checkResponse(t, recorder.Body.String())
		})
	}
}
