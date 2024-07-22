package handlers

import (
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"weblineBackend/internal/services"
)

type PromotionHandler struct {
	// promotionService is the service that handles the promotion logic
	promotionService *services.PromotionService
}

func NewPromotionHandler(promotionService *services.PromotionService) *PromotionHandler {
	return &PromotionHandler{

		promotionService: promotionService,
	}
}

// CreatePromotion creates a new promotion
func (h *PromotionHandler) CreatePromotion(w http.ResponseWriter, r *http.Request) {
	// get title, description, discount and product ID
	tagline := r.FormValue("tagline")
	mainTitle := r.FormValue("main_title")
	subTitle := r.FormValue("sub_title")
	title := r.FormValue("title")
	description := r.FormValue("description")
	discountStr := r.FormValue("discount")

	var discount float64
	if discountStr != "" {
		discount, _ = strconv.ParseFloat(discountStr, 64)
	}
	productID, err := uuid.Parse(r.FormValue("product_id"))
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	// create promotion
	promotion, err := h.promotionService.CreatePromotion(r.Context(), r, tagline, mainTitle, subTitle, title, description, discount, productID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Errorf("failed to create promotion: %w", err).Error())
		return
	}

	// respond with promotion
	RespondWithJSON(w, http.StatusOK, promotion)
}

// GetPromotions retrieves all promotions
func (h *PromotionHandler) GetPromotions(w http.ResponseWriter, r *http.Request) {
	// get promotions
	promotions, err := h.promotionService.GetPromotions(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Errorf("failed to get promotions: %w", err).Error())
		return
	}

	// respond with promotions
	RespondWithJSON(w, http.StatusOK, promotions)
}
