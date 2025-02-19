package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"weblineBackend/internal/model"
	"weblineBackend/internal/services"

	"github.com/gorilla/mux"
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
	// parse the form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse multipart form: %v", err))
		return
	}

	params := &model.CreateCategoryParams{
		Name:            r.FormValue("name"),
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
			RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve image file: %v", err))
			return
		}
	} else {
		defer file.Close()
		image = &model.ImageFile{
			File:       file,
			FileHeader: header,
		}
	}

	// create category
	err = h.categoryService.CreateCategoryService(r.Context(), params, image)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create category")
		return
	}

	// respond with category
	RespondWithSuccess(w, http.StatusCreated, "Category created successfully")
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
	parentID := mux.Vars(r)["parentId"]

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

// UploadCategoryImageHandler uploads a category image
func (h *CategoryHandler) UploadCategoryImageHandler(w http.ResponseWriter, r *http.Request) {
	// get category ID
	categoryID := r.FormValue("id")

	// upload category image
	err := h.categoryService.UpdateCategoryImageService(r.Context(), r, categoryID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to upload category image")
		return
	}

	// respond with success
	RespondWithSuccess(w, http.StatusOK, "Category image uploaded successfully")
}

// GetCollectionCategoriesHandler retrieves collection categories
func (h *CategoryHandler) GetCollectionCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	// get collection categories
	collectionCategories, err := h.categoryService.GetCollectionCategoriesService(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get collection categories")
		return
	}

	// respond with collection categories
	RespondWithJSON(w, http.StatusOK, collectionCategories)
}

// GetV2CategoryHierarchyHandler retrieves the V2 category hierarchy
func (h *CategoryHandler) GetV2CategoryHierarchyHandler(w http.ResponseWriter, r *http.Request) {
	// get V2 category hierarchy
	v2CategoryHierarchy, err := h.categoryService.GetV2CategoryHierarchy(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get V2 category hierarchy")
		return
	}

	// respond with V2 category hierarchy
	RespondWithJSON(w, http.StatusOK, v2CategoryHierarchy)
}

// DeleteCategoryHandler deletes a category
func (h *CategoryHandler) DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	// get category ID
	id := mux.Vars(r)["id"]

	// delete category
	err := h.categoryService.DeleteCategoryService(r.Context(), id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete category")
		return
	}

	// respond with success
	RespondWithSuccess(w, http.StatusOK, "Category deleted successfully")
}

// GetCategoryDetailsHandler retrieves category details
func (h *CategoryHandler) GetCategoryDetailsHandler(w http.ResponseWriter, r *http.Request) {
	// get category ID
	slug := mux.Vars(r)["slug"]

	if slug == "" {
		RespondWithError(w, http.StatusBadRequest, "Category slug is required")
		return
	}

	// get category details
	categoryDetails, err := h.categoryService.GetCategoryDetailsService(r.Context(), slug)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get category details")
		return
	}

	// respond with category details
	RespondWithJSON(w, http.StatusOK, categoryDetails)
}

// GetCategorySEOHandler retrieves category SEO
func (h *CategoryHandler) GetCategorySEOHandler(w http.ResponseWriter, r *http.Request) {
	// get category ID
	slug := mux.Vars(r)["slug"]

	if slug == "" {
		RespondWithError(w, http.StatusBadRequest, "Category slug is required")
		return
	}

	// get category SEO
	categorySEO, err := h.categoryService.GetCategorySEOService(r.Context(), slug)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get category SEO")
		return
	}

	// respond with category SEO
	RespondWithJSON(w, http.StatusOK, categorySEO)
}
