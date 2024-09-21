package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"weblineBackend/internal/model"
	"weblineBackend/internal/services"

	"github.com/gorilla/mux"
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

// DeletePromotion deletes a promotion
func (h *PromotionHandler) DeletePromotion(w http.ResponseWriter, r *http.Request) {
	// Extract the promotion slug
	slug := mux.Vars(r)["slug"]

	// Delete the promotion
	if err := h.promotionService.DeletePromotion(r.Context(), slug); err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Errorf("failed to delete promotion: %w", err).Error())
		return
	}

	// Respond with success
	RespondWithSuccess(w, http.StatusOK, "Promotion deleted successfully")
}

// DeletePromotions deletes multiple promotions
func (h *PromotionHandler) DeletePromotions(w http.ResponseWriter, r *http.Request) {
	// Extract the promotion slugs
	var params struct {
		Slugs []string `json:"slugs"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// Delete the promotions
	if err := h.promotionService.DeletePromotions(r.Context(), params.Slugs); err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Errorf("failed to delete promotions: %w", err).Error())
		return
	}

	// Respond with success
	RespondWithSuccess(w, http.StatusOK, "Promotions deleted successfully")
}

// ArchivePromotions archives multiple promotions
func (h *PromotionHandler) ArchivePromotions(w http.ResponseWriter, r *http.Request) {
	// Extract the promotion slugs
	var params struct {
		Slugs []string `json:"slugs"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// Archive the promotions
	if err := h.promotionService.ArchivePromotions(r.Context(), params.Slugs); err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Errorf("failed to archive promotions: %w", err).Error())
		return
	}

	// Respond with success
	RespondWithSuccess(w, http.StatusOK, "Promotions archived successfully")
}

// DraftPromotions drafts multiple promotions
func (h *PromotionHandler) DraftPromotions(w http.ResponseWriter, r *http.Request) {
	// Extract the promotion slugs
	var params struct {
		Slugs []string `json:"slugs"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// Draft the promotions
	if err := h.promotionService.DraftPromotions(r.Context(), params.Slugs); err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Errorf("failed to draft promotions: %w", err).Error())
		return
	}

	// Respond with success
	RespondWithSuccess(w, http.StatusOK, "Promotions drafted successfully")
}

// ActivatePromotions activates multiple promotions
func (h *PromotionHandler) ActivatePromotions(w http.ResponseWriter, r *http.Request) {
	// Extract the promotion slugs
	var params struct {
		Slugs []string `json:"slugs"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// Activate the promotions
	if err := h.promotionService.ActivatePromotions(r.Context(), params.Slugs); err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Errorf("failed to activate promotions: %w", err).Error())
		return
	}

	// Respond with success
	RespondWithSuccess(w, http.StatusOK, "Promotions activated successfully")
}