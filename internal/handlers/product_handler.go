package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"weblineBackend/internal/model"
	"weblineBackend/internal/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ProductHandler struct {
	productService    *services.ProductService
	productSEOService *services.ProductSEOService
}

func NewProductHandler(productService *services.ProductService, productSEOService *services.ProductSEOService) *ProductHandler {
	return &ProductHandler{
		productService:    productService,
		productSEOService: productSEOService,
	}
}

// CreateProductHandler creates a new product
func (h *ProductHandler) CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Stock       int32   `json:"stock"`
		CategoryID  string  `json:"category_id"`
		PartNumber  string  `json:"part_number"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// create product
	product, err := h.productService.CreateProduct(r.Context(), params.Name, params.Description, params.Price, params.CategoryID, params.PartNumber, params.Stock)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create product")
		return
	}

	// respond with product
	RespondWithJSON(w, http.StatusOK, product)
}

// GetProductByIDHandler gets a product by its ID
func (h *ProductHandler) GetProductByIDHandler(w http.ResponseWriter, r *http.Request) {
	// get product ID
	vars := mux.Vars(r)
	productID := vars["id"]

	// get product
	product, err := h.productService.GetProductByID(r.Context(), productID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get product")
		return
	}

	// respond with product
	RespondWithJSON(w, http.StatusOK, product)
}

// GetAllProductsHandler gets all products
func (h *ProductHandler) GetAllProductsHandler(w http.ResponseWriter, r *http.Request) {
	// get page and page size
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("limit")

	page, pageSize, err := GetPageAndPageSize(pageStr, pageSizeStr)

	// get filters
	categoryNamesStr := r.URL.Query().Get("category_names")
	colorsStr := r.URL.Query().Get("colors")
	processorsStr := r.URL.Query().Get("processors")
	storageStr := r.URL.Query().Get("storage")
	sizesStr := r.URL.Query().Get("sizes")
	priceFromStr := r.URL.Query().Get("price_from")
	priceToStr := r.URL.Query().Get("price_to")
	sortBy := r.URL.Query().Get("sort")

	categoryNames := parseCommaSeparatedValues(categoryNamesStr)
	colors := parseCommaSeparatedValues(colorsStr)
	processors := parseCommaSeparatedValues(processorsStr)
	storage := parseCommaSeparatedValues(storageStr)
	sizes := parseCommaSeparatedValues(sizesStr)

	priceFrom, err := parsePrice(priceFromStr, 0)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid price_from")
		return
	}

	priceTo, err := parsePrice(priceToStr, 999999)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid price_to")
		return
	}

	var products *model.PaginationResult[[]*model.Product]
	switch sortBy {
	case "price_asc":
		products, err = h.productService.GetAllProductsByFiltersPriceAsc(r.Context(), categoryNames, colors, processors, storage, sizes, priceFrom, priceTo, page, pageSize)
	case "price_desc":
		products, err = h.productService.GetAllProductsByFiltersPriceDesc(r.Context(), categoryNames, colors, processors, storage, sizes, priceFrom, priceTo, page, pageSize)
	case "name_asc":
		products, err = h.productService.GetAllProductsByFiltersNameAsc(r.Context(), categoryNames, colors, processors, storage, sizes, priceFrom, priceTo, page, pageSize)
	case "name_desc":
		products, err = h.productService.GetAllProductsByFiltersNameDesc(r.Context(), categoryNames, colors, processors, storage, sizes, priceFrom, priceTo, page, pageSize)
	case "newest":
		products, err = h.productService.GetAllProductsByFiltersNewest(r.Context(), categoryNames, colors, processors, storage, sizes, priceFrom, priceTo, page, pageSize)
	case "oldest":
		products, err = h.productService.GetAllProductsByFiltersOldest(r.Context(), categoryNames, colors, processors, storage, sizes, priceFrom, priceTo, page, pageSize)
	default:
		products, err = h.productService.GetAllProductsByFiltersNewest(r.Context(), categoryNames, colors, processors, storage, sizes, priceFrom, priceTo, page, pageSize)
	}

	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get products by filters")
		return
	}

	RespondWithJSON(w, http.StatusOK, products)
}

// UpdateProductHandler updates a product
func (h *ProductHandler) UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
	// get product ID
	vars := mux.Vars(r)
	productID := vars["id"]

	var params struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		CategoryID  string  `json:"category_id"`
		Stock       int32   `json:"stock"`
		Featured    bool    `json:"featured"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// update product
	product, err := h.productService.UpdateProduct(r.Context(), productID, params.Name, params.Description, params.Price, params.CategoryID, params.Stock, params.Featured)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update product")
		return
	}

	// respond with product
	RespondWithJSON(w, http.StatusOK, product)
}

// SoftDeleteProductHandler deletes a product
func (h *ProductHandler) SoftDeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	// get product ID
	vars := mux.Vars(r)
	productID := vars["id"]

	// delete product
	err := h.productService.SoftDeleteProduct(r.Context(), productID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete product")
		return
	}

	// respond with success
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Product deleted"})
}

// GetProductsByCategoryIDHandler gets all products by category ID
func (h *ProductHandler) GetProductsByCategoryIDHandler(w http.ResponseWriter, r *http.Request) {
	// get category ID
	vars := mux.Vars(r)
	categoryID := vars["id"]

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page, pageSize, err := GetPageAndPageSize(pageStr, pageSizeStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid page or page size: %v", err))
		return
	}

	// get products
	products, err := h.productService.GetProductsByCategoryID(r.Context(), categoryID, pageSize, page)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get products by category ID")
		return
	}

	// respond with products
	RespondWithJSON(w, http.StatusOK, products)
}

// GetProductsByParentCategoryIDHandler gets all products by parent category ID
func (h *ProductHandler) GetProductsByParentCategoryIDHandler(w http.ResponseWriter, r *http.Request) {
	// get category ID
	vars := mux.Vars(r)
	categoryID := vars["id"]

	// get products
	products, err := h.productService.GetProductsByParentCategoryID(r.Context(), categoryID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get products by parent category ID")
		return
	}

	// respond with products
	RespondWithJSON(w, http.StatusOK, products)
}

func (h *ProductHandler) GetProductsByFiltersHandler(w http.ResponseWriter, r *http.Request) {
	categoryID := mux.Vars(r)["category_id"]
	categoryNamesStr := r.URL.Query().Get("category_names")
	colorsStr := r.URL.Query().Get("colors")
	processorsStr := r.URL.Query().Get("processors")
	storageStr := r.URL.Query().Get("storage")
	sizesStr := r.URL.Query().Get("sizes")
	priceFromStr := r.URL.Query().Get("price_from")
	priceToStr := r.URL.Query().Get("price_to")
	sortBy := r.URL.Query().Get("sort")

	categoryNames := parseCommaSeparatedValues(categoryNamesStr)
	colors := parseCommaSeparatedValues(colorsStr)
	processors := parseCommaSeparatedValues(processorsStr)
	storage := parseCommaSeparatedValues(storageStr)
	sizes := parseCommaSeparatedValues(sizesStr)

	priceFrom, err := parsePrice(priceFromStr, 0)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid price_from")
		return
	}

	priceTo, err := parsePrice(priceToStr, 999999)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid price_to")
		return
	}

	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	var products []model.FilterProduct
	switch sortBy {
	case "price_asc":
		products, err = h.productService.GetProductsByFiltersPriceAsc(r.Context(), categoryUUID, categoryNames, colors, processors, storage, sizes, priceFrom, priceTo)
	case "price_desc":
		products, err = h.productService.GetProductsByFiltersPriceDesc(r.Context(), categoryUUID, categoryNames, colors, processors, storage, sizes, priceFrom, priceTo)
	case "name_asc":
		products, err = h.productService.GetProductsByFiltersNameAsc(r.Context(), categoryUUID, categoryNames, colors, processors, storage, sizes, priceFrom, priceTo)
	case "name_desc":
		products, err = h.productService.GetProductsByFiltersNameDesc(r.Context(), categoryUUID, categoryNames, colors, processors, storage, sizes, priceFrom, priceTo)
	case "newest":
		products, err = h.productService.GetProductsByFiltersNewest(r.Context(), categoryUUID, categoryNames, colors, processors, storage, sizes, priceFrom, priceTo)
	case "oldest":
		products, err = h.productService.GetProductsByFiltersOldest(r.Context(), categoryUUID, categoryNames, colors, processors, storage, sizes, priceFrom, priceTo)
	default:
		products, err = h.productService.GetProductsByFiltersDefault(r.Context(), categoryUUID, categoryNames, colors, processors, storage, sizes, priceFrom, priceTo)
	}

	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get products by filters")
		return
	}

	RespondWithJSON(w, http.StatusOK, products)
}

func parseCommaSeparatedValues(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func parsePrice(s string, defaultValue float64) (float64, error) {
	if s == "" {
		return defaultValue, nil
	}
	return strconv.ParseFloat(s, 64)
}

// SearchProductsHandler searches products by name
func (h *ProductHandler) SearchProductsHandler(w http.ResponseWriter, r *http.Request) {
	// get query
	query := r.URL.Query().Get("query")

	// get products
	products, err := h.productService.SearchProducts(r.Context(), query)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to search products")
		return
	}

	// respond with products
	RespondWithJSON(w, http.StatusOK, products)
}

// GetProductsByFilterOptionsHandler gets all products by filter options
func (h *ProductHandler) GetProductsByFilterOptionsHandler(w http.ResponseWriter, r *http.Request) {
	// get filter options
	filterOptions, err := h.productService.GetFilterOptions(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get filter options")
		return
	}

	// respond with filter options
	RespondWithJSON(w, http.StatusOK, filterOptions)
}

// GetFilterOptionsByCategoryNameHandler gets filter options by category name
func (h *ProductHandler) GetFilterOptionsByCategoryNameHandler(w http.ResponseWriter, r *http.Request) {
	// get category name
	categoryName := mux.Vars(r)["name"]

	// get filter options
	filterOptions, err := h.productService.GetFilterOptionsByCategoryName(r.Context(), categoryName)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get filter options")
		return
	}

	// respond with filter options
	RespondWithJSON(w, http.StatusOK, filterOptions)
}

// GetAllProductSitemapHandler gets all products for sitemap
func (h *ProductHandler) GetAllProductSitemapHandler(w http.ResponseWriter, r *http.Request) {
	// get products
	products, err := h.productService.GetAllProductSitemap(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get products for sitemap")
		return
	}

	// respond with products
	RespondWithJSON(w, http.StatusOK, products)
}

// GetProductSEOHandler gets the SEO information of a product
func (h *ProductHandler) GetProductSEOHandler(w http.ResponseWriter, r *http.Request) {
	// get product ID
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid slug")
		return
	}

	// get product SEO
	productSEO, err := h.productSEOService.GetProductSEO(r.Context(), slug)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			RespondWithError(w, http.StatusNotFound, "Product SEO not found")
			return
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to get product SEO")
			return
		}
	}

	// respond with product SEO
	RespondWithJSON(w, http.StatusOK, productSEO)
}

// GetProductBySlugHandler gets a product by its slug
func (h *ProductHandler) GetProductBySlugHandler(w http.ResponseWriter, r *http.Request) {
	// get product slug
	vars := mux.Vars(r)
	slug := vars["slug"]

	// get product
	product, err := h.productService.GetProductBySlug(r.Context(), slug)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			RespondWithError(w, http.StatusNotFound, "Product not found")
			return
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to get product by slug")
			return
		}
	}

	// respond with product
	RespondWithJSON(w, http.StatusOK, product)
}

// GetProductsHandler gets all products
func (h *ProductHandler) GetProductsHandler(w http.ResponseWriter, r *http.Request) {
	// get products
	products, err := h.productService.GetProducts(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get products")
		return
	}

	// respond with products
	RespondWithJSON(w, http.StatusOK, products)
}

// GetProductDetailHandler gets a product detail by its ID
func (h *ProductHandler) GetProductDetailHandler(w http.ResponseWriter, r *http.Request) {
	// get product ID
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid slug")
		return
	}

	// get product detail
	productDetail, err := h.productService.GetProductDetail(r.Context(), slug)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			RespondWithError(w, http.StatusNotFound, "Product detail not found")
			return
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to get product detail")
			return
		}
	}

	// respond with product detail
	RespondWithJSON(w, http.StatusOK, productDetail)
}
