package handlers

import (
	"encoding/json"
	"net/http"
	"weblineBackend/internal/services"

	"go.uber.org/zap"
)

type InquiryHandler struct {
	logger         *zap.Logger
	inquiryService *services.InquiryService
}

func NewInquiryHandler(logger *zap.Logger, inquiryService *services.InquiryService) *InquiryHandler {
	return &InquiryHandler{
		logger:         logger,
		inquiryService: inquiryService,
	}
}

type SubmitInquiryRequest struct {
	ProductID string `json:"productID"`
	Email     string `json:"email"`
	Message   string `json:"message"`
}

// SubmitInquiry submits a new inquiry
func (h *InquiryHandler) SubmitInquiry(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req SubmitInquiryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Submit inquiry
	err := h.inquiryService.SubmitInquiry(r.Context(), req.ProductID, req.Email, req.Message)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondWithSuccess(w, http.StatusCreated, "Inquiry submitted successfully")
}
