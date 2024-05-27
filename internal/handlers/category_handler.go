package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"weblineBackend/internal/services"
)

type CategoryHandler struct {
	categoryService *services.CategoryService
}

func NewCategoryHandler(categoryService *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// CreateCategoryHandler creates a new category
func (h *CategoryHandler) CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	// params
	var params struct {
		Name     string `json:"name"`
		ParentID string `json:"parent_id"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// create category
	category, err := h.categoryService.CreateCategoryService(r.Context(), params.Name, params.ParentID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create category")
		return
	}

	// respond with category
	RespondWithJSON(w, http.StatusOK, category)
}

// GetCategoryByIDHandler retrieves a category by its ID
func (h *CategoryHandler) GetCategoryByIDHandler(w http.ResponseWriter, r *http.Request) {
	// get category ID
	id := mux.Vars(r)["id"]

	// get category
	category, err := h.categoryService.GetCategoryByIDService(r.Context(), id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get category")
		return
	}

	// respond with category
	RespondWithJSON(w, http.StatusOK, category)
}

// GetCategoriesHandler retrieves all categories
func (h *CategoryHandler) GetCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	// get categories
	categories, err := h.categoryService.GetCategoriesService(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get categories")
		return
	}

	// respond with categories
	RespondWithJSON(w, http.StatusOK, categories)
}

// UpdateCategoryHandler updates a category
func (h *CategoryHandler) UpdateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	// get category ID
	id := mux.Vars(r)["id"]

	// params
	var params struct {
		Name     string `json:"name"`
		ParentID string `json:"parent_id"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// update category
	category, err := h.categoryService.UpdateCategoryService(r.Context(), id, params.Name, params.ParentID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update category")
		return
	}

	// respond with category
	RespondWithJSON(w, http.StatusOK, category)
}

// SoftDeleteCategoryHandler marks a category as inactive
func (h *CategoryHandler) SoftDeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	// get category ID
	id := mux.Vars(r)["id"]

	// soft delete category
	err := h.categoryService.SoftDeleteCategoryService(r.Context(), id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to soft delete category")
		return
	}

	// respond with success
	RespondWithSuccess(w, http.StatusOK, "Category soft deleted successfully")
}

// GetCategoriesWithProductsCountHandler retrieves all categories with products count
func (h *CategoryHandler) GetCategoriesWithProductsCountHandler(w http.ResponseWriter, r *http.Request) {
	// get categories with products count
	categories, err := h.categoryService.GetCategoriesWithProductsCountService(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get categories with products count")
		return
	}

	// respond with categories
	RespondWithJSON(w, http.StatusOK, categories)
}

// GetCategoryTreeHandler retrieves the category tree
func (h *CategoryHandler) GetCategoryTreeHandler(w http.ResponseWriter, r *http.Request) {
	// get category tree
	categoryTree, err := h.categoryService.GetCategoryTreeService(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get category tree")
		return
	}

	// respond with category tree
	RespondWithJSON(w, http.StatusOK, categoryTree)
}

// CheckCategoryExistenceHandler checks if a category exists
func (h *CategoryHandler) CheckCategoryExistenceHandler(w http.ResponseWriter, r *http.Request) {
	// get category ID
	id := mux.Vars(r)["id"]

	// check category existence
	exists, err := h.categoryService.CheckCategoryExistenceService(r.Context(), id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to check category existence")
		return
	}

	// respond with existence
	RespondWithJSON(w, http.StatusOK, exists)
}

// GetCategoriesWithSubcategoryCountHandler retrieves all categories with subcategory count
func (h *CategoryHandler) GetCategoriesWithSubcategoryCountHandler(w http.ResponseWriter, r *http.Request) {
	// get categories with subcategory count
	categories, err := h.categoryService.GetCategoriesWithSubcategoryCountService(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get categories with subcategory count")
		return
	}

	// respond with categories
	RespondWithJSON(w, http.StatusOK, categories)
}

// GetCategoriesByParentIDHandler retrieves categories by their parent ID
func (h *CategoryHandler) GetCategoriesByParentIDHandler(w http.ResponseWriter, r *http.Request) {
	// get parent ID
	parentID := r.URL.Query().Get("parent_id")

	// get categories
	categories, err := h.categoryService.GetCategoriesByParentIDService(r.Context(), parentID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get categories by parent ID")
		return
	}

	// respond with categories
	RespondWithJSON(w, http.StatusOK, categories)
}

// GetParentCategoriesHandler retrieves parent categories
func (h *CategoryHandler) GetParentCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	// get parent categories
	parentCategories, err := h.categoryService.GetParentCategoriesService(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get parent categories")
		return
	}

	// respond with parent categories
	RespondWithJSON(w, http.StatusOK, parentCategories)
}

// GetCategoryByNameHandler retrieves a category by its name
func (h *CategoryHandler) GetCategoryByNameHandler(w http.ResponseWriter, r *http.Request) {
	// get category name
	name := mux.Vars(r)["name"]

	// get category
	category, err := h.categoryService.GetCategoryByNameService(r.Context(), name)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get category by name")
		return
	}

	// respond with category
	RespondWithJSON(w, http.StatusOK, category)
}
