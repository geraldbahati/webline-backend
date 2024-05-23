package handlers

import (
	"encoding/json"
	"net/http"
	"weblineBackend/internal/services"
)

type ProductSpecificationHandler struct {
	productSpecificationService *services.ProductService
}

func NewProductSpecificationHandler(productSpecificationService *services.ProductService) *ProductSpecificationHandler {
	return &ProductSpecificationHandler{
		productSpecificationService: productSpecificationService,
	}
}

// CreateProductSpecificationHandler creates a new product specification
func (h *ProductSpecificationHandler) CreateProductSpecificationHandler(w http.ResponseWriter, r *http.Request) {
	// params
	var params struct {
		ProductID string `json:"product_id"`
		SpecName  string `json:"spec_name"`
		SpecValue string `json:"spec_value"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// create product specification
	productSpecification, err := h.productSpecificationService.CreateProductSpecification(r.Context(), params.ProductID, params.SpecName, params.SpecValue)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create product specification")
		return
	}

	// respond with product specification
	RespondWithJSON(w, http.StatusOK, productSpecification)
}

// ListProductSpecificationsByProductIDHandler retrieves a list of product specifications by product ID
func (h *ProductSpecificationHandler) ListProductSpecificationsByProductIDHandler(w http.ResponseWriter, r *http.Request) {
	// get product ID
	productID := r.URL.Query().Get("product_id")

	// get product specifications
	productSpecifications, err := h.productSpecificationService.ListProductSpecificationsByProductID(r.Context(), productID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get product specifications")
		return
	}

	// respond with product specifications
	RespondWithJSON(w, http.StatusOK, productSpecifications)
}

// DeleteProductSpecificationHandler deletes a product specification by its ID
func (h *ProductSpecificationHandler) DeleteProductSpecificationHandler(w http.ResponseWriter, r *http.Request) {
	// get product specification ID
	id := r.URL.Query().Get("id")

	// delete product specification
	err := h.productSpecificationService.DeleteProductSpecification(r.Context(), id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete product specification")
		return
	}

	// respond with success
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Product specification deleted successfully"})
}
