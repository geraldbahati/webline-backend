package handlers

import (
	"encoding/json"
	"net/http"
	"weblineBackend/internal/services"
)

type ProductOptionHandler struct {
	productOptionService *services.ProductService
}

func NewProductOptionHandler(productOptionService *services.ProductService) *ProductOptionHandler {
	return &ProductOptionHandler{productOptionService: productOptionService}
}

// CreateProductOptionHandler creates a new product option
func (h *ProductOptionHandler) CreateProductOptionHandler(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// create product option
	productOption, err := h.productOptionService.CreateProductOption(r.Context(), params.Name, params.Value)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create product option")
		return
	}

	// respond with product option
	RespondWithJSON(w, http.StatusOK, productOption)
}

// ListProductOptionsByProductIDHandler lists all product options by product ID
func (h *ProductOptionHandler) ListProductOptionsByProductIDHandler(w http.ResponseWriter, r *http.Request) {
	// get product ID
	productID := r.URL.Query().Get("product_id")

	// list product options
	productOptions, err := h.productOptionService.ListProductOptionsByProductID(r.Context(), productID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to list product options")
		return
	}

	// respond with product options
	RespondWithJSON(w, http.StatusOK, productOptions)
}

// UpdateProductOptionHandler updates a product option by its ID
func (h *ProductOptionHandler) UpdateProductOptionHandler(w http.ResponseWriter, r *http.Request) {
	var params struct {
		OptionName string `json:"option_name"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// get product option ID
	productOptionID := r.URL.Query().Get("id")

	// update product option
	productOption, err := h.productOptionService.UpdateProductOption(r.Context(), productOptionID, params.OptionName)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update product option")
		return
	}

	// respond with product option
	RespondWithJSON(w, http.StatusOK, productOption)
}

// DeleteProductOptionHandler deletes a product option by its ID
func (h *ProductOptionHandler) DeleteProductOptionHandler(w http.ResponseWriter, r *http.Request) {
	// get product option ID
	productOptionID := r.URL.Query().Get("id")

	// delete product option
	err := h.productOptionService.DeleteProductOption(r.Context(), productOptionID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete product option")
		return
	}

	// respond with success
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Product option deleted successfully"})
}

// CreateProductOptionValueHandler creates a new product option value
func (h *ProductOptionHandler) CreateProductOptionValueHandler(w http.ResponseWriter, r *http.Request) {
	var params struct {
		OptionID        string  `json:"option_id"`
		OptionValue     string  `json:"option_value"`
		AdditionalPrice float64 `json:"additional_price"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// create product option value
	productOptionValue, err := h.productOptionService.CreateProductOptionValue(r.Context(), params.OptionID, params.OptionValue, params.AdditionalPrice)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create product option value")
		return
	}

	// respond with product option value
	RespondWithJSON(w, http.StatusOK, productOptionValue)
}

// ListProductOptionValuesByOptionIDHandler lists all product option values by option ID
func (h *ProductOptionHandler) ListProductOptionValuesByOptionIDHandler(w http.ResponseWriter, r *http.Request) {
	// get option ID
	optionID := r.URL.Query().Get("option_id")

	// list product option values
	productOptionValues, err := h.productOptionService.ListProductOptionValuesByOptionID(r.Context(), optionID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to list product option values")
		return
	}

	// respond with product option values
	RespondWithJSON(w, http.StatusOK, productOptionValues)
}

// UpdateProductOptionValueHandler updates a product option value by its ID
func (h *ProductOptionHandler) UpdateProductOptionValueHandler(w http.ResponseWriter, r *http.Request) {
	var params struct {
		OptionValue     string  `json:"option_value"`
		AdditionalPrice float64 `json:"additional_price"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// get product option value ID
	productOptionValueID := r.URL.Query().Get("id")

	// update product option value
	productOptionValue, err := h.productOptionService.UpdateProductOptionValue(r.Context(), productOptionValueID, params.OptionValue, params.AdditionalPrice)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update product option value")
		return
	}

	// respond with product option value
	RespondWithJSON(w, http.StatusOK, productOptionValue)
}

// DeleteProductOptionValueHandler deletes a product option value by its ID
func (h *ProductOptionHandler) DeleteProductOptionValueHandler(w http.ResponseWriter, r *http.Request) {
	// get product option value ID
	productOptionValueID := r.URL.Query().Get("id")

	// delete product option value
	err := h.productOptionService.DeleteProductOptionValue(r.Context(), productOptionValueID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete product option value")
		return
	}

	// respond with success
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Product option value deleted successfully"})
}
