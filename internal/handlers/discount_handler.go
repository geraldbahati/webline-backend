package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"time"
	"weblineBackend/internal/services"
)

type DiscountHandler struct {
	discountService *services.DiscountService
}

func NewDiscountHandler(discountService *services.DiscountService) *DiscountHandler {
	return &DiscountHandler{
		discountService: discountService,
	}
}

// CreateDiscountHandler creates a new discount
func (h *DiscountHandler) CreateDiscountHandler(w http.ResponseWriter, r *http.Request) {
	// params
	var params struct {
		ProductID string  `json:"product_id"`
		Discount  float64 `json:"discount"`
		StartDate string  `json:"start_date"`
		EndDate   string  `json:"end_date"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// validate request body return json with specific messages
	messageObj := make(map[string]string)

	if params.ProductID == "" {
		messageObj["product_id"] = "Product ID is required"
	}
	if params.Discount == 0 {
		messageObj["discount"] = "Discount is required"

	}
	if params.StartDate == "" {
		messageObj["start_date"] = "Start date is required"
	}
	if params.EndDate == "" {
		messageObj["end_date"] = "End date is required"
	}
	if len(messageObj) > 0 {
		RespondWithJSON(w, http.StatusBadRequest, messageObj)
		return
	}

	// parse product ID
	productUUID, err := uuid.Parse(params.ProductID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	// parse start date and end date
	startDate, err := parseDate(params.StartDate)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid start date")
		return
	}

	endDate, err := parseDate(params.EndDate)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid end date")
		return
	}

	// create discount
	err = h.discountService.CreateDiscount(r.Context(), productUUID, params.Discount, startDate, endDate)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create discount")
		return
	}

	// respond with success
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Discount created successfully"})
}

// validate request body
func validateRequestBody(params struct {
	ProductID string  `json:"product_id"`
	Discount  float64 `json:"discount"`
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
}) map[string]string {
	var messageObj map[string]string
	if params.ProductID == "" {
		messageObj["product_id"] = "Product ID is required"
	}
	if params.Discount == 0 {
		messageObj["discount"] = "Discount is required"

	}
	if params.StartDate == "" {
		messageObj["start_date"] = "Start date is required"
	}
	if params.EndDate == "" {
		messageObj["end_date"] = "End date is required"
	}
	return messageObj
}

// parse date
func parseDate(date string) (time.Time, error) {
	return time.Parse("2006-01-02", date)
}
