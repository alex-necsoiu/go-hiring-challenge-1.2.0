package categories

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

// CategoryRepository defines the contract for category data access operations.
type CategoryRepository interface {
	// GetAllCategories returns all categories from the database.
	GetAllCategories() ([]models.Category, error)

	// CreateCategory persists a new category to the database.
	// Returns an error if the create operation fails (e.g., duplicate code).
	CreateCategory(category *models.Category) error
}

// CategoryItem represents a category in the response.
type CategoryItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// CategoriesListData represents the data payload for the categories list response.
type CategoriesListData struct {
	Categories []CategoryItem `json:"categories"`
}

// CategoriesResponse wraps the categories list in the response envelope.
type CategoriesResponse struct {
	Data CategoriesListData `json:"data"`
}

// CategoryResponse wraps a single category in the response envelope.
type CategoryResponse struct {
	Data CategoryItem `json:"data"`
}

// CategoriesHandler handles HTTP requests for categories.
type CategoriesHandler struct {
	repo CategoryRepository
}

// NewCategoriesHandler creates a new CategoriesHandler with the given repository.
func NewCategoriesHandler(r CategoryRepository) *CategoriesHandler {
	return &CategoriesHandler{
		repo: r,
	}
}

// HandleGet returns all categories.
func (h *CategoriesHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.GetAllCategories()
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "failed to fetch categories")
		return
	}

	// Build response
	categoryItems := make([]CategoryItem, len(categories))
	for i, c := range categories {
		categoryItems[i] = CategoryItem{
			Code: c.Code,
			Name: c.Name,
		}
	}

	response := CategoriesResponse{
		Data: CategoriesListData{
			Categories: categoryItems,
		},
	}

	api.OKResponse(w, response)
}

// HandleCreate creates a new category.
func (h *CategoriesHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var req struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if strings.TrimSpace(req.Code) == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "code is required")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "name is required")
		return
	}

	// Create category
	category := &models.Category{
		Code: req.Code,
		Name: req.Name,
	}

	if err := h.repo.CreateCategory(category); err != nil {
		// Check if it's a duplicate key error
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique") {
			api.ErrorResponse(w, http.StatusConflict, "category code already exists")
			return
		}
		api.ErrorResponse(w, http.StatusInternalServerError, "failed to create category")
		return
	}

	// Return created category
	response := CategoryResponse{
		Data: CategoryItem{
			Code: category.Code,
			Name: category.Name,
		},
	}

	api.OKResponse(w, response)
}
