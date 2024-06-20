package handlers

import (
	"encoding/json"
	"net/http"
	"weblineBackend/internal/services"

	"github.com/gorilla/mux"
)

type ProductColorHandler struct {
	productColorService *services.ProductService
}

func NewProductColorHandler(productColorService *services.ProductService) *ProductColorHandler {
	return &ProductColorHandler{productColorService: productColorService}
}

// CreateProductColorHandler creates a new product color
func (h *ProductColorHandler) CreateProductColorHandler(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ProductID string `json:"product_id"`
		Color     string `json:"color"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// create product color
	if _, err := h.productColorService.CreateProductColor(r.Context(), params.ProductID, params.Color); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create product color")
		return
	}

	// respond with success
	RespondWithSuccess(w, http.StatusCreated, "Product color created successfully")
}

// ListProductColorsByProductIDHandler lists all product colors by product ID
func (h *ProductColorHandler) ListProductColorsByProductIDHandler(w http.ResponseWriter, r *http.Request) {
	// get product ID
	productID := r.URL.Query().Get("product_id")

	// list product colors
	productColors, err := h.productColorService.ListProductColorsByProductID(r.Context(), productID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to list product colors")
		return
	}

	// respond with product colors
	RespondWithJSON(w, http.StatusOK, productColors)
}

// UpdateProductColorHandler updates a product color
func (h *ProductColorHandler) UpdateProductColorHandler(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ProductID string `json:"product_id"`
		Color     string `json:"color"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// update product color
	if _, err := h.productColorService.UpdateProductColor(r.Context(), params.ProductID, params.Color); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update product color")
		return
	}

	// respond with success
	RespondWithSuccess(w, http.StatusOK, "Product color updated successfully")
}

// DeleteProductColorHandler deletes a product color
func (h *ProductColorHandler) DeleteProductColorHandler(w http.ResponseWriter, r *http.Request) {
	productID := mux.Vars(r)["id"]

	// delete product color
	if err := h.productColorService.DeleteProductColor(r.Context(), productID); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete product color")
		return
	}

	// respond with success
	RespondWithSuccess(w, http.StatusOK, "Product color deleted successfully")
}
