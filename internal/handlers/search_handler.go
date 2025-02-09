package handlers

import (
	"net/http"
	"strconv"

	"weblineBackend/internal/services"

	"go.uber.org/zap"
)

// SearchHandler provides HTTP endpoints for searching products.
type SearchHandler struct {
	searchService services.SearchService
	logger        *zap.Logger
}

// NewSearchHandler creates a new instance of SearchHandler.
func NewSearchHandler(searchService services.SearchService, logger *zap.Logger) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
		logger:        logger,
	}
}

// SearchProducts handles a non-paginated search request.
// Expects the query parameter "q" and optional "limit" (default 10).
func (h *SearchHandler) SearchProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		RespondWithError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	limit := 10
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	results, err := h.searchService.SearchProducts(r.Context(), q, limit)
	if err != nil {
		h.logger.Error("SearchProducts error", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to search products")
		return
	}

	RespondWithJSON(w, http.StatusOK, results)
}

// SearchProductsPaginated handles a paginated search request.
// Expects "q" for query, and optional "page" and "limit" query parameters.
func (h *SearchHandler) SearchProductsPaginated(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		RespondWithError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	page := 1
	if pStr := r.URL.Query().Get("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	paginatedResult, err := h.searchService.SearchProductsPaginated(r.Context(), q, page, limit)
	if err != nil {
		h.logger.Error("SearchProductsPaginated error", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to search products paginated")
		return
	}

	RespondWithJSON(w, http.StatusOK, paginatedResult)
}

// AutocompleteSuggestions handles requests for autocomplete suggestions.
// Expects "q" and an optional "limit" (default is 5).
func (h *SearchHandler) AutocompleteSuggestions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		RespondWithError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	limit := 5
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	suggestions, err := h.searchService.AutocompleteSuggestions(r.Context(), q, limit)
	if err != nil {
		h.logger.Error("AutocompleteSuggestions error", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to get autocomplete suggestions")
		return
	}

	RespondWithJSON(w, http.StatusOK, suggestions)
}
