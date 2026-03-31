package categories

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/mytheresa/go-hiring-challenge/models"
)

// mockCategoryRepository is a test double implementing CategoryRepository interface
type mockCategoryRepository struct {
	getAllCategoriesFunc func() ([]models.Category, error)
	createCategoryFunc   func(category *models.Category) error
}

func (m *mockCategoryRepository) GetAllCategories() ([]models.Category, error) {
	return m.getAllCategoriesFunc()
}

func (m *mockCategoryRepository) CreateCategory(category *models.Category) error {
	return m.createCategoryFunc(category)
}

func TestHandleGet(t *testing.T) {
	tests := []struct {
		name           string
		mockFunc       func() ([]models.Category, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name: "happy path - returns 200 with categories list",
			mockFunc: func() ([]models.Category, error) {
				return []models.Category{
					{
						ID:   1,
						Code: "clothing",
						Name: "Clothing",
					},
					{
						ID:   2,
						Code: "shoes",
						Name: "Shoes",
					},
					{
						ID:   3,
						Code: "accessories",
						Name: "Accessories",
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp CategoriesResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, 3, len(resp.Data.Categories))
				assert.Equal(t, "clothing", resp.Data.Categories[0].Code)
				assert.Equal(t, "Clothing", resp.Data.Categories[0].Name)
				assert.Equal(t, "shoes", resp.Data.Categories[1].Code)
				assert.Equal(t, "accessories", resp.Data.Categories[2].Code)
			},
		},
		{
			name: "empty list - returns 200",
			mockFunc: func() ([]models.Category, error) {
				return []models.Category{}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp CategoriesResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, 0, len(resp.Data.Categories))
			},
		},
		{
			name: "repository error - returns 500",
			mockFunc: func() ([]models.Category, error) {
				return nil, assert.AnError
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp["error"], "")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCategoryRepository{
				getAllCategoriesFunc: tt.mockFunc,
			}

			handler := &CategoriesHandler{repo: mock}
			req := httptest.NewRequest("GET", "/categories", nil)
			recorder := httptest.NewRecorder()

			handler.HandleGet(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			tt.checkResponse(t, recorder.Body.String())
		})
	}
}

func TestHandleCreate(t *testing.T) {
	tests := []struct {
		name              string
		requestBody       interface{}
		mockFunc          func(category *models.Category) error
		expectedStatus    int
		checkResponse     func(t *testing.T, body string)
		checkCategoryData func(t *testing.T) // for checking what was passed to mock
	}{
		{
			name: "successful creation - returns 200",
			requestBody: map[string]string{
				"code": "new-category",
				"name": "New Category",
			},
			mockFunc: func(category *models.Category) error {
				category.ID = 4
				return nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp CategoryResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "new-category", resp.Data.Code)
				assert.Equal(t, "New Category", resp.Data.Name)
			},
			checkCategoryData: func(t *testing.T) {},
		},
		{
			name: "missing code - returns 400",
			requestBody: map[string]string{
				"name": "New Category",
			},
			mockFunc:       func(category *models.Category) error { return nil },
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp["error"], "code")
			},
			checkCategoryData: func(t *testing.T) {},
		},
		{
			name: "missing name - returns 400",
			requestBody: map[string]string{
				"code": "new-category",
			},
			mockFunc:       func(category *models.Category) error { return nil },
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp["error"], "name")
			},
			checkCategoryData: func(t *testing.T) {},
		},
		{
			name:           "invalid JSON body - returns 400",
			requestBody:    "invalid json",
			mockFunc:       func(category *models.Category) error { return nil },
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp["error"])
			},
			checkCategoryData: func(t *testing.T) {},
		},
		{
			name: "duplicate code (unique constraint) - returns 409",
			requestBody: map[string]string{
				"code": "clothing",
				"name": "Clothing Duplicate",
			},
			mockFunc: func(category *models.Category) error {
				// Simulate unique constraint violation
				return errors.New("UNIQUE constraint failed: categories.code")
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp["error"], "already exists")
			},
			checkCategoryData: func(t *testing.T) {},
		},
		{
			name: "repository error - returns 500",
			requestBody: map[string]string{
				"code": "new-category",
				"name": "New Category",
			},
			mockFunc: func(category *models.Category) error {
				return assert.AnError
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp["error"])
			},
			checkCategoryData: func(t *testing.T) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			if str, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tt.requestBody)
			}

			mock := &mockCategoryRepository{
				createCategoryFunc: tt.mockFunc,
			}

			handler := &CategoriesHandler{repo: mock}
			req := httptest.NewRequest("POST", "/categories", bytes.NewReader(bodyBytes))
			recorder := httptest.NewRecorder()

			handler.HandleCreate(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			tt.checkResponse(t, recorder.Body.String())
			tt.checkCategoryData(t)
		})
	}
}
