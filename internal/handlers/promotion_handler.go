package handlers

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"
	"weblineBackend/internal/model"
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

// GetV2Promotions retrieves all promotions for dashboard
func (h *PromotionHandler) GetV2Promotions(w http.ResponseWriter, r *http.Request) {
	// get promotions
	promotions, err := h.promotionService.GetV2Promotions(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Errorf("failed to get promotions: %w", err).Error())
		return
	}

	// respond with promotions
	RespondWithJSON(w, http.StatusOK, promotions)
}

// CreateOrEditV2Promotion creates or edits a promotion
func (h *PromotionHandler) CreateOrEditV2Promotion(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse multipart form: %v", err))
		return
	}

	// Extract form values with a helper function
	params := model.CreatePromotionParams{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Slug:        r.FormValue("slug"),
		Status:      r.FormValue("status"),
	}

	// Parse and validate the start and end dates
	if startDate, endDate, err := parsePromotionDates(r); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	} else {
		params.StartDate = startDate
		params.EndDate = endDate
	}

	// Extract product slugs
	if slugs, ok := r.MultipartForm.Value["productSlugs"]; ok {
		params.ProductSlugs = slugs
	} else {
		params.ProductSlugs = []string{}
	}

	// Handle the image file upload
	file, header, err := r.FormFile("image")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			RespondWithError(w, http.StatusBadRequest, "Image file is required")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve image file: %v", err))
		return
	}
	defer file.Close()

	image := &model.ImageFile{
		File:       file,
		FileHeader: header,
	}

	// Create or edit the promotion
	if err := h.promotionService.CreateOrEditV2Promotion(r.Context(), &params, image); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create or edit promotion")
		return
	}

	// Respond with success
	RespondWithSuccess(w, http.StatusOK, "Promotion created or edited successfully")
}

// parsePromotionDates parses and validates the start and end dates from the form
func parsePromotionDates(r *http.Request) (startDate, endDate time.Time, err error) {
	startDateStr := r.FormValue("startDate")
	endDateStr := r.FormValue("endDate")

	startDate, err = time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		return startDate, endDate, fmt.Errorf("invalid start date: %v", err)
	}

	endDate, err = time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		return startDate, endDate, fmt.Errorf("invalid end date: %v", err)
	}

	return startDate, endDate, nil
}

// GetPromotionDetails retrieves the details of a promotion
func (h *PromotionHandler) GetPromotionDetails(w http.ResponseWriter, r *http.Request) {
	// Extract the promotion slug
	slug := mux.Vars(r)["slug"]

	// Get the promotion details
	promotion, err := h.promotionService.GetPromotionDetails(r.Context(), slug)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Errorf("failed to get promotion details: %w", err).Error())
		return
	}

	// Respond with the promotion details
	RespondWithJSON(w, http.StatusOK, promotion)
}
