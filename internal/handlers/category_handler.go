package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"weblineBackend/internal/model"
	"weblineBackend/internal/services"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Define error constants if app_errors package is not available
var (
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized access")
	ErrInvalidInput = errors.New("invalid input")
	ErrDuplicate    = errors.New("duplicate resource")
	ErrInUse        = errors.New("resource in use")
)

type CategoryHandler struct {
	categoryService *services.CategoryService
	logger          *zap.Logger
}

// NewCategoryHandler creates a new CategoryHandler with improved dependency injection
func NewCategoryHandler(categoryService *services.CategoryService, logger *zap.Logger) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
		logger:          logger,
	}
}

// CreateCategoryHandler creates a new category with improved error handling
func (h *CategoryHandler) CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Parse the form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		h.logger.Error("failed to parse multipart form", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse multipart form: %v", err))
		return
	}

	// Validate required fields
	name := r.FormValue("name")
	if name == "" {
		RespondWithError(w, http.StatusBadRequest, "Category name is required")
		return
	}

	params := &model.CreateCategoryParams{
		Name:            name,
		Description:     r.FormValue("description"),
		MetaTitle:       r.FormValue("metaTitle"),
		MetaDescription: r.FormValue("metaDescription"),
		ParentID:        r.FormValue("categoryID"),
		Slug:            r.FormValue("slug"),
	}

	// Handle the image file upload (the image is optional and will be automatically optimized if provided)
	var image *model.ImageFile
	file, header, err := r.FormFile("image")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			// No image provided; image remains nil.
			image = nil
		} else {
			h.logger.Error("failed to retrieve image file", zap.Error(err))
			RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve image file: %v", err))
			return
		}
	} else {
		defer file.Close()
		image = &model.ImageFile{
			File:       file,
			FileHeader: header,
		}
	}

	// Create category
	err = h.categoryService.CreateCategoryService(ctx, params, image)
	if err != nil {
		h.logger.Error("failed to create category", zap.Error(err), zap.String("name", name))

		// Return appropriate status code based on error type
		if errors.Is(err, ErrInvalidInput) {
			RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		} else if errors.Is(err, ErrUnauthorized) {
			RespondWithError(w, http.StatusUnauthorized, "Unauthorized: Admin privileges required")
			return
		} else if errors.Is(err, ErrDuplicate) {
			RespondWithError(w, http.StatusConflict, "Category with this name or slug already exists")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to create category")
		return
	}

	// Respond with success
	RespondWithSuccess(w, http.StatusCreated, "Category created successfully")
}

// GetCategoryByIDHandler retrieves a category by its ID with caching
func (h *CategoryHandler) GetCategoryByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Get category ID from URL
	id := mux.Vars(r)["id"]
	if id == "" {
		RespondWithError(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	// Get category
	category, err := h.categoryService.GetCategoryByIDService(ctx, id)
	if err != nil {
		h.logger.Error("failed to get category by ID", zap.Error(err), zap.String("categoryID", id))

		if errors.Is(err, ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Category not found")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to get category")
		return
	}

	// Set caching headers (5 minutes cache for GET requests)
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("ETag", fmt.Sprintf("W/\"%s\"", id))

	// Respond with category
	RespondWithJSON(w, http.StatusOK, category)
}

// GetCategoriesHandler retrieves all categories with optimized caching
func (h *CategoryHandler) GetCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Check for query parameters
	activeOnlyStr := r.URL.Query().Get("active_only")
	var activeOnly bool = true
	if activeOnlyStr != "" {
		var err error
		activeOnly, err = strconv.ParseBool(activeOnlyStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid active_only parameter")
			return
		}
	}

	// Get categories - pass activeOnly parameter to the service layer
	categories, err := h.categoryService.GetCategoriesServiceWithFilter(ctx, activeOnly)
	if err != nil {
		h.logger.Error("failed to get categories", zap.Error(err), zap.Bool("activeOnly", activeOnly))

		if ctx.Err() != nil {
			RespondWithError(w, http.StatusGatewayTimeout, "Request timeout")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to get categories")
		return
	}

	// Set caching headers (5 minutes cache for GET requests)
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("Vary", "Accept-Encoding, Origin")

	// Generate ETag based on data size
	etag := fmt.Sprintf("W/\"%d\"", len(categories))
	w.Header().Set("ETag", etag)

	// Check If-None-Match header for 304 response
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Respond with categories
	RespondWithJSON(w, http.StatusOK, categories)
}

// SoftDeleteCategoryHandler marks a category as inactive with improved validation
func (h *CategoryHandler) SoftDeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Get category ID from URL
	id := mux.Vars(r)["id"]
	if id == "" {
		RespondWithError(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	// Soft delete category
	err := h.categoryService.SoftDeleteCategoryService(ctx, id)
	if err != nil {
		h.logger.Error("failed to soft delete category", zap.Error(err), zap.String("categoryID", id))

		if errors.Is(err, ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Category not found")
			return
		} else if errors.Is(err, ErrUnauthorized) {
			RespondWithError(w, http.StatusUnauthorized, "Unauthorized: Admin privileges required")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to soft delete category")
		return
	}

	// Respond with success
	RespondWithSuccess(w, http.StatusOK, "Category soft deleted successfully")
}

// GetCategoriesWithProductsCountHandler retrieves all categories with product counts
func (h *CategoryHandler) GetCategoriesWithProductsCountHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Get categories with products count
	categories, err := h.categoryService.GetCategoriesWithProductsCountService(ctx)
	if err != nil {
		h.logger.Error("failed to get categories with products count", zap.Error(err))

		if ctx.Err() != nil {
			RespondWithError(w, http.StatusGatewayTimeout, "Request timeout")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to get categories with products count")
		return
	}

	// Set caching headers (2 minutes cache for this data as it may change more frequently)
	w.Header().Set("Cache-Control", "public, max-age=120")
	w.Header().Set("Vary", "Accept-Encoding, Origin")

	// Respond with categories
	RespondWithJSON(w, http.StatusOK, categories)
}

// GetCategoryTreeHandler retrieves the category tree with enhanced caching
func (h *CategoryHandler) GetCategoryTreeHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Get category tree
	categoryTree, err := h.categoryService.GetCategoryTreeService(ctx)
	if err != nil {
		h.logger.Error("failed to get category tree", zap.Error(err))

		if ctx.Err() != nil {
			RespondWithError(w, http.StatusGatewayTimeout, "Request timeout")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to get category tree")
		return
	}

	// Set caching headers (10 minutes cache for tree data as it's more stable)
	w.Header().Set("Cache-Control", "public, max-age=600")
	w.Header().Set("Vary", "Accept-Encoding, Origin")

	// Generate an ETag
	etag := fmt.Sprintf("W/\"%d\"", len(categoryTree))
	w.Header().Set("ETag", etag)

	// Check If-None-Match header for 304 response
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Respond with category tree
	RespondWithJSON(w, http.StatusOK, categoryTree)
}

// CheckCategoryExistenceHandler checks if a category exists
func (h *CategoryHandler) CheckCategoryExistenceHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Get category ID from URL
	id := mux.Vars(r)["id"]
	if id == "" {
		RespondWithError(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	// Check category existence
	exists, err := h.categoryService.CheckCategoryExistenceService(ctx, id)
	if err != nil {
		h.logger.Error("failed to check category existence", zap.Error(err), zap.String("categoryID", id))
		RespondWithError(w, http.StatusInternalServerError, "Failed to check category existence")
		return
	}

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Set caching headers (5 minutes)
	w.Header().Set("Cache-Control", "public, max-age=300")

	// Respond with existence status (no body needed for HEAD request)
	w.WriteHeader(http.StatusOK)
}

// GetCategoriesWithSubcategoryCountHandler retrieves all categories with subcategory count
func (h *CategoryHandler) GetCategoriesWithSubcategoryCountHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Get categories with subcategory count
	categories, err := h.categoryService.GetCategoriesWithSubcategoryCountService(ctx)
	if err != nil {
		h.logger.Error("failed to get categories with subcategory count", zap.Error(err))

		if ctx.Err() != nil {
			RespondWithError(w, http.StatusGatewayTimeout, "Request timeout")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to get categories with subcategory count")
		return
	}

	// Set caching headers
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("Vary", "Accept-Encoding, Origin")

	// Respond with categories
	RespondWithJSON(w, http.StatusOK, categories)
}

// GetCategoriesByParentIDHandler retrieves categories by their parent ID
func (h *CategoryHandler) GetCategoriesByParentIDHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Get parent ID from URL
	parentID := mux.Vars(r)["parentId"]
	if parentID == "" {
		RespondWithError(w, http.StatusBadRequest, "Parent ID is required")
		return
	}

	// Get categories
	categories, err := h.categoryService.GetCategoriesByParentIDService(ctx, parentID)
	if err != nil {
		h.logger.Error("failed to get categories by parent ID", zap.Error(err), zap.String("parentID", parentID))

		if errors.Is(err, ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Parent category not found")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to get categories by parent ID")
		return
	}

	// Set caching headers
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("Vary", "Accept-Encoding, Origin")

	// Respond with categories
	RespondWithJSON(w, http.StatusOK, categories)
}

// GetParentCategoriesHandler retrieves parent categories
func (h *CategoryHandler) GetParentCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Get parent categories
	parentCategories, err := h.categoryService.GetParentCategoriesService(ctx)
	if err != nil {
		h.logger.Error("failed to get parent categories", zap.Error(err))

		if ctx.Err() != nil {
			RespondWithError(w, http.StatusGatewayTimeout, "Request timeout")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to get parent categories")
		return
	}

	// Set caching headers (longer cache duration as parent categories change less frequently)
	w.Header().Set("Cache-Control", "public, max-age=600")
	w.Header().Set("Vary", "Accept-Encoding, Origin")

	// Respond with parent categories
	RespondWithJSON(w, http.StatusOK, parentCategories)
}

// GetCategoryByNameHandler retrieves a category by its name
func (h *CategoryHandler) GetCategoryByNameHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Get category name from URL
	name := mux.Vars(r)["name"]
	if name == "" {
		RespondWithError(w, http.StatusBadRequest, "Category name is required")
		return
	}

	// Get category
	category, err := h.categoryService.GetCategoryByNameService(ctx, name)
	if err != nil {
		h.logger.Error("failed to get category by name", zap.Error(err), zap.String("name", name))

		if errors.Is(err, ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Category not found")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to get category by name")
		return
	}

	// Set caching headers
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("Vary", "Accept-Encoding, Origin")

	// Respond with category
	RespondWithJSON(w, http.StatusOK, category)
}

// UploadCategoryImageHandler uploads a category image
func (h *CategoryHandler) UploadCategoryImageHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout (longer for file uploads)
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		h.logger.Error("failed to parse multipart form", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse multipart form: %v", err))
		return
	}

	// Get category ID
	categoryID := r.FormValue("id")
	if categoryID == "" {
		RespondWithError(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	// Upload category image
	err := h.categoryService.UpdateCategoryImageService(ctx, r, categoryID)
	if err != nil {
		h.logger.Error("failed to upload category image", zap.Error(err), zap.String("categoryID", categoryID))

		if errors.Is(err, ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Category not found")
			return
		} else if errors.Is(err, ErrUnauthorized) {
			RespondWithError(w, http.StatusUnauthorized, "Unauthorized: Admin privileges required")
			return
		} else if errors.Is(err, ErrInvalidInput) {
			RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to upload category image")
		return
	}

	// Respond with success
	RespondWithSuccess(w, http.StatusOK, "Category image uploaded successfully")
}

// GetCollectionCategoriesHandler retrieves collection categories
func (h *CategoryHandler) GetCollectionCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Get collection categories
	collectionCategories, err := h.categoryService.GetCollectionCategoriesService(ctx)
	if err != nil {
		h.logger.Error("failed to get collection categories", zap.Error(err))

		if ctx.Err() != nil {
			RespondWithError(w, http.StatusGatewayTimeout, "Request timeout")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to get collection categories")
		return
	}

	// Set caching headers
	w.Header().Set("Cache-Control", "public, max-age=600")
	w.Header().Set("Vary", "Accept-Encoding, Origin")

	// Respond with collection categories
	RespondWithJSON(w, http.StatusOK, collectionCategories)
}

// GetV2CategoryHierarchyHandler retrieves the V2 category hierarchy with efficient caching
func (h *CategoryHandler) GetV2CategoryHierarchyHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Check for query parameters
	activeOnlyStr := r.URL.Query().Get("active_only")
	activeOnly := true
	if activeOnlyStr != "" {
		var err error
		activeOnly, err = strconv.ParseBool(activeOnlyStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid active_only parameter")
			return
		}
	}

	// get V2 category hierarchy
	v2CategoryHierarchy, err := h.categoryService.GetV2CategoryHierarchyWithFilter(ctx, activeOnly)
	if err != nil {
		h.logger.Error("failed to get V2 category hierarchy", zap.Error(err))

		if ctx.Err() != nil {
			RespondWithError(w, http.StatusGatewayTimeout, "Request timeout")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to get V2 category hierarchy")
		return
	}

	// Set caching headers (10 minutes for hierarchy which changes less frequently)
	w.Header().Set("Cache-Control", "public, max-age=600")
	w.Header().Set("Vary", "Accept-Encoding, Origin, X-Requested-With")

	// Generate an ETag based on data size and active_only parameter
	etag := fmt.Sprintf("W/\"%d-%v\"", len(v2CategoryHierarchy), activeOnly)
	w.Header().Set("ETag", etag)

	// Check If-None-Match header for 304 response
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// respond with V2 category hierarchy
	RespondWithJSON(w, http.StatusOK, v2CategoryHierarchy)
}

// DeleteCategoryHandler deletes a category
func (h *CategoryHandler) DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Get category ID from URL
	id := mux.Vars(r)["id"]
	if id == "" {
		RespondWithError(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	// Delete category
	err := h.categoryService.DeleteCategoryService(ctx, id)
	if err != nil {
		h.logger.Error("failed to delete category", zap.Error(err), zap.String("categoryID", id))

		if errors.Is(err, ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Category not found")
			return
		} else if errors.Is(err, ErrUnauthorized) {
			RespondWithError(w, http.StatusUnauthorized, "Unauthorized: Admin privileges required")
			return
		} else if errors.Is(err, ErrInUse) {
			RespondWithError(w, http.StatusConflict, "Cannot delete: Category is in use by products")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to delete category")
		return
	}

	// Respond with success
	RespondWithSuccess(w, http.StatusOK, "Category deleted successfully")
}

// GetCategoryDetailsHandler retrieves category details
func (h *CategoryHandler) GetCategoryDetailsHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Get category slug from URL
	slug := mux.Vars(r)["slug"]
	if slug == "" {
		RespondWithError(w, http.StatusBadRequest, "Category slug is required")
		return
	}

	// Get category details
	categoryDetails, err := h.categoryService.GetCategoryDetailsService(ctx, slug)
	if err != nil {
		h.logger.Error("failed to get category details", zap.Error(err), zap.String("slug", slug))

		if errors.Is(err, ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Category not found")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to get category details")
		return
	}

	// Set caching headers
	w.Header().Set("Cache-Control", "public, max-age=600")
	w.Header().Set("Vary", "Accept-Encoding, Origin")

	// Respond with category details
	RespondWithJSON(w, http.StatusOK, categoryDetails)
}

// GetCategorySEOHandler retrieves category SEO
func (h *CategoryHandler) GetCategorySEOHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Get category slug from URL
	slug := mux.Vars(r)["slug"]
	if slug == "" {
		RespondWithError(w, http.StatusBadRequest, "Category slug is required")
		return
	}

	// Get category SEO
	categorySEO, err := h.categoryService.GetCategorySEOService(ctx, slug)
	if err != nil {
		h.logger.Error("failed to get category SEO", zap.Error(err), zap.String("slug", slug))

		if errors.Is(err, ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Category not found")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to get category SEO")
		return
	}

	// Set caching headers (longer for SEO data which changes rarely)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Vary", "Accept-Encoding, Origin")

	// Respond with category SEO
	RespondWithJSON(w, http.StatusOK, categorySEO)
}

// GetCategoryHierarchyStatsHandler retrieves category hierarchy with stats
func (h *CategoryHandler) GetCategoryHierarchyStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Get category ID from URL
	id := mux.Vars(r)["id"]
	if id == "" {
		RespondWithError(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	// Check for query parameters
	activeOnlyStr := r.URL.Query().Get("active_only")
	activeOnly := true
	if activeOnlyStr != "" {
		var err error
		activeOnly, err = strconv.ParseBool(activeOnlyStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid active_only parameter")
			return
		}
	}

	// Get hierarchy stats
	stats, err := h.categoryService.GetCategoryHierarchyStatsService(ctx, id, activeOnly)
	if err != nil {
		h.logger.Error("failed to get category hierarchy stats",
			zap.Error(err),
			zap.String("categoryID", id),
			zap.Bool("activeOnly", activeOnly))

		if errors.Is(err, ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Category not found")
			return
		} else if ctx.Err() != nil {
			RespondWithError(w, http.StatusGatewayTimeout, "Request timeout")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to get category hierarchy stats")
		return
	}

	// Set caching headers
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("Vary", "Accept-Encoding, Origin")

	// Generate an ETag based on category ID and active_only
	etag := fmt.Sprintf("W/\"%s-%v\"", id, activeOnly)
	w.Header().Set("ETag", etag)

	// Respond with hierarchy stats
	RespondWithJSON(w, http.StatusOK, stats)
}

// GetDirectChildrenWithStatsHandler retrieves immediate children with their stats
func (h *CategoryHandler) GetDirectChildrenWithStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Get category slug from URL
	slug := mux.Vars(r)["slug"]
	if slug == "" {
		RespondWithError(w, http.StatusBadRequest, "Category slug is required")
		return
	}

	// Check for query parameters
	activeOnlyStr := r.URL.Query().Get("active_only")
	activeOnly := true
	if activeOnlyStr != "" {
		var err error
		activeOnly, err = strconv.ParseBool(activeOnlyStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid active_only parameter")
			return
		}
	}

	// Get direct children with stats
	children, err := h.categoryService.GetDirectChildrenWithStatsService(ctx, slug, activeOnly)
	if err != nil {
		h.logger.Error("failed to get direct children with stats",
			zap.Error(err),
			zap.String("categorySlug", slug),
			zap.Bool("activeOnly", activeOnly))

		if errors.Is(err, ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Category not found")
			return
		} else if ctx.Err() != nil {
			RespondWithError(w, http.StatusGatewayTimeout, "Request timeout")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to get direct children with stats")
		return
	}

	// Set caching headers
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("Vary", "Accept-Encoding, Origin")

	// Respond with children
	RespondWithJSON(w, http.StatusOK, children)
}

// BatchUpdateCategoryPositionsHandler updates positions for multiple categories
func (h *CategoryHandler) BatchUpdateCategoryPositionsHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Parse request body
	var updates map[string]int
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate updates
	if len(updates) == 0 {
		RespondWithError(w, http.StatusBadRequest, "No updates provided")
		return
	}

	// Perform batch update
	err := h.categoryService.BatchUpdateCategoryPositionsService(ctx, updates)
	if err != nil {
		h.logger.Error("failed to batch update category positions",
			zap.Error(err),
			zap.Int("count", len(updates)))

		if errors.Is(err, ErrUnauthorized) {
			RespondWithError(w, http.StatusUnauthorized, "Unauthorized: Admin privileges required")
			return
		} else if errors.Is(err, ErrInvalidInput) {
			RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to update category positions")
		return
	}

	// Respond with success
	RespondWithSuccess(w, http.StatusOK, "Category positions updated successfully")
}

// GetPopularCategoriesHandler retrieves popular categories
func (h *CategoryHandler) GetPopularCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	// Add request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	// Get limit from query parameters (default to 10)
	limitStr := r.URL.Query().Get("limit")
	limit := int32(10)
	if limitStr != "" {
		limitVal, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil || limitVal < 1 || limitVal > 50 {
			RespondWithError(w, http.StatusBadRequest, "Invalid limit parameter (must be between 1 and 50)")
			return
		}
		limit = int32(limitVal)
	}

	// Get popular categories
	categories, err := h.categoryService.GetPopularCategoriesService(ctx, limit)
	if err != nil {
		h.logger.Error("failed to get popular categories",
			zap.Error(err),
			zap.Int32("limit", limit))

		if ctx.Err() != nil {
			RespondWithError(w, http.StatusGatewayTimeout, "Request timeout")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to get popular categories")
		return
	}

	// Set caching headers
	w.Header().Set("Cache-Control", "public, max-age=1800") // 30 minutes
	w.Header().Set("Vary", "Accept-Encoding, Origin")

	// Respond with categories
	RespondWithJSON(w, http.StatusOK, categories)
}
