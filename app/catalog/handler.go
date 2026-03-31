package catalog

import (
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ProductRepository defines the contract for product data access operations.
type ProductRepository interface {
	// GetProducts returns a paginated list of products matching the given filter.
	// Returns the products, total count, and an error if the query fails.
	GetProducts(filter models.ProductFilter) ([]models.Product, int64, error)

	// GetProductByCode retrieves a single product by its code.
	// Returns the product or an error if not found or query fails.
	GetProductByCode(code string) (*models.Product, error)
}

// CategoryItem represents a category in the response.
type CategoryItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// ProductItem represents a product in the catalog list response.
type ProductItem struct {
	Code     string       `json:"code"`
	Name     string       `json:"name"`
	Price    string       `json:"price"`
	Category CategoryItem `json:"category"`
}

// VariantItem represents a variant in the product detail response.
type VariantItem struct {
	Name  string `json:"name"`
	SKU   string `json:"sku"`
	Price string `json:"price"`
}

// ProductDetail represents the full product details with variants.
type ProductDetail struct {
	Code     string        `json:"code"`
	Name     string        `json:"name"`
	Price    string        `json:"price"`
	Category CategoryItem  `json:"category"`
	Variants []VariantItem `json:"variants"`
}

// CatalogListData represents the data payload for the products list response.
type CatalogListData struct {
	Products []ProductItem `json:"products"`
	Total    int64         `json:"total"`
	Offset   int           `json:"offset"`
	Limit    int           `json:"limit"`
}

// CatalogResponse wraps the catalog list data in the response envelope.
type CatalogResponse struct {
	Data CatalogListData `json:"data"`
}

// ProductDetailResponse wraps the product detail in the response envelope.
type ProductDetailResponse struct {
	Data ProductDetail `json:"data"`
}

// CatalogHandler handles HTTP requests for the product catalog.
type CatalogHandler struct {
	repo ProductRepository
}

// NewCatalogHandler creates a new CatalogHandler with the given repository.
func NewCatalogHandler(r ProductRepository) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

// HandleGet returns a paginated list of products with optional filtering.
func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	// Parse and validate pagination parameters
	offset, err := parseIntParam(r, "offset", 0)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "invalid offset parameter")
		return
	}
	if offset < 0 {
		api.ErrorResponse(w, http.StatusBadRequest, "offset must be >= 0")
		return
	}

	limit, err := parseIntParam(r, "limit", 10)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "invalid limit parameter")
		return
	}
	if limit < 1 || limit > 100 {
		api.ErrorResponse(w, http.StatusBadRequest, "limit must be between 1 and 100")
		return
	}

	// Parse optional filters
	categoryCode := r.URL.Query().Get("category")

	maxPrice := decimal.Zero
	priceStr := r.URL.Query().Get("priceLessThan")
	if priceStr != "" {
		price, err := decimal.NewFromString(priceStr)
		if err != nil {
			api.ErrorResponse(w, http.StatusBadRequest, "invalid priceLessThan parameter")
			return
		}
		maxPrice = price
	}

	// Query products from repository
	filter := models.ProductFilter{
		Offset:       offset,
		Limit:        limit,
		CategoryCode: categoryCode,
		MaxPrice:     maxPrice,
	}

	products, total, err := h.repo.GetProducts(filter)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "failed to fetch products")
		return
	}

	// Build response
	productItems := make([]ProductItem, len(products))
	for i, p := range products {
		productItems[i] = ProductItem{
			Code:  p.Code,
			Name:  p.Name,
			Price: p.Price.String(),
			Category: CategoryItem{
				Code: p.Category.Code,
				Name: p.Category.Name,
			},
		}
	}

	response := CatalogResponse{
		Data: CatalogListData{
			Products: productItems,
			Total:    total,
			Offset:   offset,
			Limit:    limit,
		},
	}

	api.OKResponse(w, response)
}

// HandleGetByCode returns the details of a single product by its code.
func (h *CatalogHandler) HandleGetByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "product code is required")
		return
	}

	product, err := h.repo.GetProductByCode(code)
	if err == gorm.ErrRecordNotFound {
		api.ErrorResponse(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "failed to fetch product")
		return
	}

	// Build variant items with price inheritance
	variantItems := make([]VariantItem, len(product.Variants))
	for i, v := range product.Variants {
		variantPrice := v.Price
		// If variant price is zero, inherit from product price
		if v.Price.IsZero() {
			variantPrice = product.Price
		}
		variantItems[i] = VariantItem{
			Name:  v.Name,
			SKU:   v.SKU,
			Price: variantPrice.String(),
		}
	}

	response := ProductDetailResponse{
		Data: ProductDetail{
			Code:  product.Code,
			Name:  product.Name,
			Price: product.Price.String(),
			Category: CategoryItem{
				Code: product.Category.Code,
				Name: product.Category.Name,
			},
			Variants: variantItems,
		},
	}

	api.OKResponse(w, response)
}

// parseIntParam parses an integer query parameter, returning the default if not provided.
func parseIntParam(r *http.Request, paramName string, defaultVal int) (int, error) {
	paramStr := r.URL.Query().Get(paramName)
	if paramStr == "" {
		return defaultVal, nil
	}
	return strconv.Atoi(paramStr)
}

