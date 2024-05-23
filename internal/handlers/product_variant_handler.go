package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"weblineBackend/internal/services"
)

type ProductVariantHandler struct {
	productVariantService *services.ProductService
}

func NewProductVariantHandler(productVariantService *services.ProductService) *ProductVariantHandler {
	return &ProductVariantHandler{
		productVariantService: productVariantService,
	}
}

// CreateProductVariantHandler creates a new product variant
func (h *ProductVariantHandler) CreateProductVariantHandler(w http.ResponseWriter, r *http.Request) {
	// params
	var params struct {
		ProductID       string  `json:"product_id"`
		VariantName     string  `json:"variant_name"`
		VariantValue    string  `json:"variant_value"`
		AdditionalPrice float64 `json:"additional_price"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// create product variant
	productVariant, err := h.productVariantService.CreateProductVariant(r.Context(), params.ProductID, params.VariantName, params.VariantValue, params.AdditionalPrice)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create product variant")
		return
	}

	// respond with product variant
	RespondWithJSON(w, http.StatusOK, productVariant)
}

// GetProductVariantByIDHandler retrieves a product variant by its ID
func (h *ProductVariantHandler) GetProductVariantByIDHandler(w http.ResponseWriter, r *http.Request) {
	// get product variant ID
	id := mux.Vars(r)["id"]

	// get product variant
	productVariant, err := h.productVariantService.GetProductVariantByID(r.Context(), id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get product variant")
		return
	}

	// respond with product variant
	RespondWithJSON(w, http.StatusOK, productVariant)
}

// ListProductVariantsByProductIDHandler retrieves a product variant by its product ID
func (h *ProductVariantHandler) ListProductVariantsByProductIDHandler(w http.ResponseWriter, r *http.Request) {
	// get product variant ID
	id := mux.Vars(r)["id"]

	// get product variant
	productVariant, err := h.productVariantService.ListProductVariantsByProductID(r.Context(), id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get product variant")
		return
	}

	// respond with product variant
	RespondWithJSON(w, http.StatusOK, productVariant)
}

// UpdateProductVariantHandler updates a product variant
func (h *ProductVariantHandler) UpdateProductVariantHandler(w http.ResponseWriter, r *http.Request) {
	// get product variant ID
	id := mux.Vars(r)["id"]

	var params struct {
		VariantName     string  `json:"variant_name"`
		VariantValue    string  `json:"variant_value"`
		AdditionalPrice float64 `json:"additional_price"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// update product variant
	productVariant, err := h.productVariantService.UpdateProductVariant(r.Context(), id, params.VariantName, params.VariantValue, params.AdditionalPrice)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update product variant")
		return
	}

	// respond with product variant
	RespondWithJSON(w, http.StatusOK, productVariant)
}

// DeleteProductVariantHandler deletes a product variant
func (h *ProductVariantHandler) DeleteProductVariantHandler(w http.ResponseWriter, r *http.Request) {
	// get product variant ID
	id := mux.Vars(r)["id"]

	// delete product variant
	err := h.productVariantService.DeleteProductVariant(r.Context(), id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete product variant")
		return
	}

	// respond with success
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{"message": "Product variant deleted successfully"})
}
