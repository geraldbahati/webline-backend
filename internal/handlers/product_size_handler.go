package handlers

import (
	"encoding/json"
	"net/http"
	"weblineBackend/internal/services"

	"github.com/gorilla/mux"
)

type ProductSizeHandler struct {
	productSizeService *services.ProductSizeService
}

func NewProductSizeHandler(productSizeService *services.ProductSizeService) *ProductSizeHandler {
	return &ProductSizeHandler{productSizeService: productSizeService}
}

// CreateProductSizeHandler creates a new product size
func (h *ProductSizeHandler) CreateProductSizeHandler(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ProductID       string  `json:"product_id"`
		Size            string  `json:"size"`
		AdditionalPrice float64 `json:"additional_price"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// create product size
	if err := h.productSizeService.CreateProductSize(r.Context(), params.ProductID, params.Size, params.AdditionalPrice); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create product size")
		return
	}

	// respond with success
	RespondWithSuccess(w, http.StatusCreated, "Product size created successfully")
}

// ListProductSizesByProductIDHandler lists all product sizes by product ID
func (h *ProductSizeHandler) ListProductSizesByProductIDHandler(w http.ResponseWriter, r *http.Request) {
	// get product ID
	productID := r.URL.Query().Get("product_id")

	// list product sizes
	productSizes, err := h.productSizeService.ListProductSizesByProductID(r.Context(), productID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to list product sizes")
		return
	}

	// respond with product sizes
	RespondWithJSON(w, http.StatusOK, productSizes)
}

// UpdateProductSizeHandler updates a product size
func (h *ProductSizeHandler) UpdateProductSizeHandler(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ProductID       string  `json:"product_id"`
		Size            string  `json:"size"`
		AdditionalPrice float64 `json:"additional_price"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// update product size
	if err := h.productSizeService.UpdateProductSize(r.Context(), params.ProductID, params.Size, params.AdditionalPrice); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update product size")
		return
	}

	// respond with success
	RespondWithSuccess(w, http.StatusOK, "Product size updated successfully")
}

// DeleteProductSizeHandler deletes a product size
func (h *ProductSizeHandler) DeleteProductSizeHandler(w http.ResponseWriter, r *http.Request) {
	productID := mux.Vars(r)["id"]

	// delete product size
	if err := h.productSizeService.DeleteProductSize(r.Context(), productID); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete product size")
		return
	}

	// respond with success
	RespondWithSuccess(w, http.StatusOK, "Product size deleted successfully")
}
