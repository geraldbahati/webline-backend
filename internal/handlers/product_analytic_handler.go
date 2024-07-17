package handlers

import (
	"net/http"
	"strconv"
	"weblineBackend/internal/services"
)

type ProductAnalyticHandler struct {
	// productAnalyticService is the service that handles the product analytic logic
	productAnalyticService *services.ProductAnalyticService
}

func NewProductAnalyticHandler(productAnalyticService *services.ProductAnalyticService) *ProductAnalyticHandler {
	return &ProductAnalyticHandler{
		productAnalyticService: productAnalyticService,
	}
}

// GetBestSellerProducts retrieves the best seller products
func (h *ProductAnalyticHandler) GetBestSellerProducts(w http.ResponseWriter, r *http.Request) {
	limit := r.URL.Query().Get("limit")

	// get limit
	var limitInt int32
	if limit != "" {
		i, err := strconv.ParseInt(limit, 10, 32)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid limit")
			return
		}

		limitInt = int32(i)
	} else {
		limitInt = 20
	}

	// get best seller products
	bestSellerProducts, err := h.productAnalyticService.GetBestSellerProducts(r.Context(), limitInt)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get best seller products")
		return
	}

	// respond with products
	RespondWithJSON(w, http.StatusOK, bestSellerProducts)
}

// GetFeaturedProducts retrieves the featured products
func (h *ProductAnalyticHandler) GetFeaturedProducts(w http.ResponseWriter, r *http.Request) {
	limit := r.URL.Query().Get("limit")

	// get limit
	var limitInt int32
	if limit != "" {
		i, err := strconv.ParseInt(limit, 10, 32)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid limit")
			return
		}

		limitInt = int32(i)
	} else {
		limitInt = 20
	}

	// get featured products
	featuredProducts, err := h.productAnalyticService.GetFeaturedProducts(r.Context(), limitInt)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get featured products")
		return
	}

	// respond with products
	RespondWithJSON(w, http.StatusOK, featuredProducts)
}

// GetNewArrivalProducts retrieves the new arrival products
func (h *ProductAnalyticHandler) GetNewArrivalProducts(w http.ResponseWriter, r *http.Request) {
	limit := r.URL.Query().Get("limit")

	// get limit
	var limitInt int32
	if limit != "" {
		i, err := strconv.ParseInt(limit, 10, 32)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid limit")
			return
		}

		limitInt = int32(i)
	} else {
		limitInt = 20
	}

	// get new arrival products
	newArrivalProducts, err := h.productAnalyticService.GetNewArrivalProducts(r.Context(), limitInt)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get new arrival products")
		return
	}

	// respond with products
	RespondWithJSON(w, http.StatusOK, newArrivalProducts)
}

// GetDailyDealsProducts retrieves the daily deals products
func (h *ProductAnalyticHandler) GetDailyDealsProducts(w http.ResponseWriter, r *http.Request) {
	// get daily deals products
	dailyDealsProducts, err := h.productAnalyticService.GetDailyDealsProducts(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get daily deals products")
		return
	}

	// respond with products
	RespondWithJSON(w, http.StatusOK, dailyDealsProducts)
}
