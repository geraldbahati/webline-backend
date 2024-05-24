package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"weblineBackend/internal/services"
)

type ProductHandler struct {
	productService *services.ProductService
}

func NewProductHandler(productService *services.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
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
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// create product
	product, err := h.productService.CreateProduct(r.Context(), params.Name, params.Description, params.Price, params.CategoryID, params.Stock)
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
	pageSizeStr := r.URL.Query().Get("page_size")

	page, pageSize, err := GetPageAndPageSize(pageStr, pageSizeStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid page or page size")
		return
	}

	// get products
	products, err := h.productService.ListProducts(r.Context(), pageSize, page)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get products")
		return
	}

	// respond with products
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

	// get products
	products, err := h.productService.GetProductsByCategoryID(r.Context(), categoryID)
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
