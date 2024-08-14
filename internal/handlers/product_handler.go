package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"weblineBackend/internal/model"
	"weblineBackend/internal/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ProductHandler struct {
	productService    *services.ProductService
	productSEOService *services.ProductSEOService
}

func NewProductHandler(productService *services.ProductService, productSEOService *services.ProductSEOService) *ProductHandler {
	return &ProductHandler{
		productService:    productService,
		productSEOService: productSEOService,
	}
}

//// GetAllProductsHandler gets all products
//func (h *ProductHandler) GetAllProductsHandler(w http.ResponseWriter, r *http.Request) {
//	// get page and page size
//	pageStr := r.URL.Query().Get("page")
//	pageSizeStr := r.URL.Query().Get("limit")
//
//	page, pageSize, err := GetPageAndPageSize(pageStr, pageSizeStr)
//
//	// get filters
//	categoryNamesStr := r.URL.Query().Get("category_names")
//	colorsStr := r.URL.Query().Get("colors")
//	processorsStr := r.URL.Query().Get("processors")
//	storageStr := r.URL.Query().Get("storage")
//	sizesStr := r.URL.Query().Get("sizes")
//	priceFromStr := r.URL.Query().Get("price_from")
//	priceToStr := r.URL.Query().Get("price_to")
//	sortBy := r.URL.Query().Get("sort")
//
//	categoryNames := parseCommaSeparatedValues(categoryNamesStr)
//	colors := parseCommaSeparatedValues(colorsStr)
//	processors := parseCommaSeparatedValues(processorsStr)
//	storage := parseCommaSeparatedValues(storageStr)
//	sizes := parseCommaSeparatedValues(sizesStr)
//
//	priceFrom, err := parsePrice(priceFromStr, 0)
//	if err != nil {
//		RespondWithError(w, http.StatusBadRequest, "Invalid price_from")
//		return
//	}
//
//	priceTo, err := parsePrice(priceToStr, 999999)
//	if err != nil {
//		RespondWithError(w, http.StatusBadRequest, "Invalid price_to")
//		return
//	}
//	//
//	//var products *model.PaginationResult[[]*model.Product]
//	//switch sortBy {
//	//case "price_asc":
//	//	products, err = h.productService.GetAllProductsByFiltersPriceAsc(r.Context(), categoryNames, colors, processors, storage, sizes, priceFrom, priceTo, page, pageSize)
//	//case "price_desc":
//	//	products, err = h.productService.GetAllProductsByFiltersPriceDesc(r.Context(), categoryNames, colors, processors, storage, sizes, priceFrom, priceTo, page, pageSize)
//	//case "name_asc":
//	//	products, err = h.productService.GetAllProductsByFiltersNameAsc(r.Context(), categoryNames, colors, processors, storage, sizes, priceFrom, priceTo, page, pageSize)
//	//case "name_desc":
//	//	products, err = h.productService.GetAllProductsByFiltersNameDesc(r.Context(), categoryNames, colors, processors, storage, sizes, priceFrom, priceTo, page, pageSize)
//	//case "newest":
//	//	products, err = h.productService.GetAllProductsByFiltersNewest(r.Context(), categoryNames, colors, processors, storage, sizes, priceFrom, priceTo, page, pageSize)
//	//case "oldest":
//	//	products, err = h.productService.GetAllProductsByFiltersOldest(r.Context(), categoryNames, colors, processors, storage, sizes, priceFrom, priceTo, page, pageSize)
//	//default:
//	//	products, err = h.productService.GetAllProductsByFiltersNewest(r.Context(), categoryNames, colors, processors, storage, sizes, priceFrom, priceTo, page, pageSize)
//	//}
//
//	//products, err := h.productService.GetAllProductsByFilters(r.Context(), categoryNames, colors, processors, storage, sizes, priceFrom, priceTo, page, pageSize, sortBy)
//
//	if err != nil {
//		RespondWithError(w, http.StatusInternalServerError, "Failed to get products by filters")
//		return
//	}
//
//	//RespondWithJSON(w, http.StatusOK, products)
//}

// GetProductsByCategoryIDHandler gets all products by category ID
func (h *ProductHandler) GetProductsByCategoryIDHandler(w http.ResponseWriter, r *http.Request) {
	// get category ID
	vars := mux.Vars(r)
	categoryID := vars["id"]

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page, pageSize, err := GetPageAndPageSize(pageStr, pageSizeStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid page or page size: %v", err))
		return
	}

	// get products
	products, err := h.productService.GetProductsByCategoryID(r.Context(), categoryID, pageSize, page)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get products by category ID")
		return
	}

	// respond with products
	RespondWithJSON(w, http.StatusOK, products)
}

//// GetProductsByParentCategoryIDHandler gets all products by parent category ID
//func (h *ProductHandler) GetProductsByParentCategoryIDHandler(w http.ResponseWriter, r *http.Request) {
//	// get category ID
//	vars := mux.Vars(r)
//	categoryID := vars["id"]
//
//	// get products
//	products, err := h.productService.GetProductsByParentCategoryID(r.Context(), categoryID)
//	if err != nil {
//		RespondWithError(w, http.StatusInternalServerError, "Failed to get products by parent category ID")
//		return
//	}
//
//	// respond with products
//	RespondWithJSON(w, http.StatusOK, products)
//}

//func (h *ProductHandler) GetProductsByFiltersHandler(w http.ResponseWriter, r *http.Request) {
//	categoryID := mux.Vars(r)["category_id"]
//	categoryNamesStr := r.URL.Query().Get("category_names")
//	colorsStr := r.URL.Query().Get("colors")
//	processorsStr := r.URL.Query().Get("processors")
//	storageStr := r.URL.Query().Get("storage")
//	sizesStr := r.URL.Query().Get("sizes")
//	priceFromStr := r.URL.Query().Get("price_from")
//	priceToStr := r.URL.Query().Get("price_to")
//	sortBy := r.URL.Query().Get("sort")
//
//	categoryNames := parseCommaSeparatedValues(categoryNamesStr)
//	colors := parseCommaSeparatedValues(colorsStr)
//	processors := parseCommaSeparatedValues(processorsStr)
//	storage := parseCommaSeparatedValues(storageStr)
//	sizes := parseCommaSeparatedValues(sizesStr)
//
//	priceFrom, err := parsePrice(priceFromStr, 0)
//	if err != nil {
//		RespondWithError(w, http.StatusBadRequest, "Invalid price_from")
//		return
//	}
//
//	priceTo, err := parsePrice(priceToStr, 999999)
//	if err != nil {
//		RespondWithError(w, http.StatusBadRequest, "Invalid price_to")
//		return
//	}
//
//	categoryUUID, err := uuid.Parse(categoryID)
//	if err != nil {
//		RespondWithError(w, http.StatusBadRequest, "Invalid category ID")
//		return
//	}
//
//	products, err := h.productService.GetProductsByFilters(
//		r.Context(),
//		categoryUUID,
//		categoryNames,
//		colors,
//		processors,
//		storage,
//		sizes,
//		priceFrom,
//		priceTo,
//		sortBy,
//	)
//
//	if err != nil {
//		RespondWithError(w, http.StatusInternalServerError, "Failed to get products by filters")
//		return
//	}
//
//	RespondWithJSON(w, http.StatusOK, products)
//}

func parseCommaSeparatedValues(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func parsePrice(s string, defaultValue float64) (float64, error) {
	if s == "" {
		return defaultValue, nil
	}
	return strconv.ParseFloat(s, 64)
}

// SearchProductsHandler searches products by name
func (h *ProductHandler) SearchProductsHandler(w http.ResponseWriter, r *http.Request) {
	// get query
	query := r.URL.Query().Get("query")

	// get products
	products, err := h.productService.SearchProducts(r.Context(), query)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to search products")
		return
	}

	// respond with products
	RespondWithJSON(w, http.StatusOK, products)
}

//// GetProductsByFilterOptionsHandler gets all products by filter options
//func (h *ProductHandler) GetProductsByFilterOptionsHandler(w http.ResponseWriter, r *http.Request) {
//	// get filter options
//	filterOptions, err := h.productService.GetFilterOptions(r.Context())
//	if err != nil {
//		RespondWithError(w, http.StatusInternalServerError, "Failed to get filter options")
//		return
//	}
//
//	// respond with filter options
//	RespondWithJSON(w, http.StatusOK, filterOptions)
//}
//
//// GetFilterOptionsByCategoryNameHandler gets filter options by category name
//func (h *ProductHandler) GetFilterOptionsByCategoryNameHandler(w http.ResponseWriter, r *http.Request) {
//	// get category name
//	categoryName := mux.Vars(r)["name"]
//
//	// get filter options
//	filterOptions, err := h.productService.GetFilterOptionsByCategoryName(r.Context(), categoryName)
//	if err != nil {
//		RespondWithError(w, http.StatusInternalServerError, "Failed to get filter options")
//		return
//	}
//
//	// respond with filter options
//	RespondWithJSON(w, http.StatusOK, filterOptions)
//}

// GetAllProductSitemapHandler gets all products for sitemap
func (h *ProductHandler) GetAllProductSitemapHandler(w http.ResponseWriter, r *http.Request) {
	// get products
	products, err := h.productService.GetAllProductSitemap(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get products for sitemap")
		return
	}

	// respond with products
	RespondWithJSON(w, http.StatusOK, products)
}

// GetProductSEOHandler gets the SEO information of a product
func (h *ProductHandler) GetProductSEOHandler(w http.ResponseWriter, r *http.Request) {
	// get product ID
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid slug")
		return
	}

	// get product SEO
	productSEO, err := h.productSEOService.GetProductSEO(r.Context(), slug)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			RespondWithError(w, http.StatusNotFound, "Product SEO not found")
			return
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to get product SEO")
			return
		}
	}

	// respond with product SEO
	RespondWithJSON(w, http.StatusOK, productSEO)
}

// GetProductBySlugHandler gets a product by its slug
func (h *ProductHandler) GetProductBySlugHandler(w http.ResponseWriter, r *http.Request) {
	// get product slug
	vars := mux.Vars(r)
	slug := vars["slug"]

	// get product
	product, err := h.productService.GetProductBySlug(r.Context(), slug)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			RespondWithError(w, http.StatusNotFound, "Product not found")
			return
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to get product by slug")
			return
		}
	}

	// respond with product
	RespondWithJSON(w, http.StatusOK, product)
}

// GetProductsHandler gets all products
func (h *ProductHandler) GetProductsHandler(w http.ResponseWriter, r *http.Request) {
	// get products
	products, err := h.productService.GetProducts(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get products")
		return
	}

	// respond with products
	RespondWithJSON(w, http.StatusOK, products)
}

// GetProductDetailHandler gets a product detail by its ID
func (h *ProductHandler) GetProductDetailHandler(w http.ResponseWriter, r *http.Request) {
	// get product ID
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid slug")
		return
	}

	// get product detail
	productDetail, err := h.productService.GetProductDetail(r.Context(), slug)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			RespondWithError(w, http.StatusNotFound, "Product detail not found")
			return
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to get product detail")
			return
		}
	}

	// respond with product detail
	RespondWithJSON(w, http.StatusOK, productDetail)
}

const maxUploadSize = 10 << 20 // 10 MB

// CreateV2ProductHandler creates a new product
func (h *ProductHandler) CreateV2ProductHandler(w http.ResponseWriter, r *http.Request) {
	// parse the form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse multipart form: %v", err))
		return
	}

	// get the form data
	var params model.CreateProductRequest
	params.Slug = r.FormValue("slug")
	params.Name = r.FormValue("name")
	params.Description = r.FormValue("description")
	priceStr := r.FormValue("price")
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid price")
		return
	}
	params.Price = price
	stockStr := r.FormValue("stock")
	stock, err := strconv.Atoi(stockStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid stock")
		return
	}
	params.Stock = stock
	categoryIDStr := r.FormValue("categoryID")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}
	params.CategoryID = categoryID
	params.Status = r.FormValue("status")
	params.PartNumber = r.FormValue("partNumber")

	// get the images
	images := r.MultipartForm.File["images"]

	// Process image Urls
	imageUrls := make([]string, 0)
	for _, imageUrl := range r.MultipartForm.Value["url"] {
		imageUrls = append(imageUrls, imageUrl)
	}
	params.ImageUrls = imageUrls

	// Process specifications
	specifications := make([]model.Specification, 0)
	for _, spec := range r.MultipartForm.Value["specifications"] {
		var specification model.Specification
		if err := json.Unmarshal([]byte(spec), &specification); err != nil {
			http.Error(w, "Failed to parse specifications", http.StatusBadRequest)
			return
		}
		specifications = append(specifications, specification)
	}
	params.Specifications = specifications

	log.Println("params: ", params)

	//create product
	err = h.productService.CreateV2Product(r.Context(), &params, images)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create product")
		return
	}

	//respond with product
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Product created/updated successfully"})
}

// DeleteProductHandler deletes a product
func (h *ProductHandler) DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	// get product slug
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid slug")
		return
	}

	// delete product
	err := h.productService.DeleteProduct(r.Context(), slug)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete product")
		return
	}

	// respond with success
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Product deleted"})
}

// ArchiveProductHandler archives a product
func (h *ProductHandler) ArchiveProductHandler(w http.ResponseWriter, r *http.Request) {
	// get product slug
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid slug")
		return
	}

	// archive product
	err := h.productService.ArchiveProduct(r.Context(), slug)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to archive product")
		return
	}

	// respond with success
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Product archived"})
}

// ArchiveProductsHandler archives multiple products
func (h *ProductHandler) ArchiveProductsHandler(w http.ResponseWriter, r *http.Request) {
	// get product slugs
	var params struct {
		Slugs []string `json:"slugs"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// archive products
	err := h.productService.ArchiveProducts(r.Context(), params.Slugs)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to archive products")
		return
	}

	// respond with success
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Products archived"})
}

// ActivateProductsHandler activates multiple products
func (h *ProductHandler) ActivateProductsHandler(w http.ResponseWriter, r *http.Request) {
	// get product slugs
	var params struct {
		Slugs []string `json:"slugs"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// activate products
	err := h.productService.ActivateProducts(r.Context(), params.Slugs)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to activate products")
		return
	}

	// respond with success
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Products activated"})
}

// DraftProductsHandler drafts multiple products
func (h *ProductHandler) DraftProductsHandler(w http.ResponseWriter, r *http.Request) {
	// get product slugs
	var params struct {
		Slugs []string `json:"slugs"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// draft products
	err := h.productService.DraftProducts(r.Context(), params.Slugs)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to draft products")
		return
	}

	// respond with success
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Products drafted"})
}

// DeleteProductsHandler deletes multiple products
func (h *ProductHandler) DeleteProductsHandler(w http.ResponseWriter, r *http.Request) {
	// get product slugs
	var params struct {
		Slugs []string `json:"slugs"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// delete products
	err := h.productService.DeleteProducts(r.Context(), params.Slugs)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete products")
		return
	}

	// respond with success
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Products deleted"})
}

// GetProductImagesBySlugHandler gets product images by its slug
func (h *ProductHandler) GetProductImagesBySlugHandler(w http.ResponseWriter, r *http.Request) {
	// get product slug
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid slug")
		return
	}

	// get product images
	productImages, err := h.productService.GetProductImagesBySlug(r.Context(), slug)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			RespondWithError(w, http.StatusNotFound, "Product images not found")
			return
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to get product images")
			return
		}
	}

	// respond with product images
	RespondWithJSON(w, http.StatusOK, productImages)
}

// GetProductPricingBySlugHandler gets product pricing by its slug
func (h *ProductHandler) GetProductPricingBySlugHandler(w http.ResponseWriter, r *http.Request) {
	// get product slug
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid slug")
		return
	}

	// get product pricing
	productPricing, err := h.productService.GetProductPricingBySlug(r.Context(), slug)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			RespondWithError(w, http.StatusNotFound, "Product pricing not found")
			return
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to get product pricing")
			return
		}
	}

	// respond with product pricing
	RespondWithJSON(w, http.StatusOK, productPricing)
}

// GetProductSpecsBySlugHandler gets product specs by its slug
func (h *ProductHandler) GetProductSpecsBySlugHandler(w http.ResponseWriter, r *http.Request) {
	// get product slug
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid slug")
		return
	}

	// get product specs
	productSpecs, err := h.productService.GetProductSpecsBySlug(r.Context(), slug)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			RespondWithError(w, http.StatusNotFound, "Product specs not found")
			return
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to get product specs")
			return
		}
	}

	// respond with product specs
	RespondWithJSON(w, http.StatusOK, productSpecs)
}

//// GetProductMetaFieldsByCategoryIDHandler gets product meta fields by category ID
//func (h *ProductHandler) GetProductMetaFieldsByCategoryIDHandler(w http.ResponseWriter, r *http.Request) {
//	// get category ID
//	vars := mux.Vars(r)
//	categoryID := vars["categoryID"]
//
//	// parse category ID
//	categoryUUID, err := uuid.Parse(categoryID)
//	if err != nil {
//		RespondWithError(w, http.StatusBadRequest, "Invalid category ID")
//		return
//	}
//
//	// get product meta fields
//	productMetaFields, err := h.productService.GetFilterOptionsByCategoryID(r.Context(), categoryUUID)
//	if err != nil {
//		RespondWithError(w, http.StatusInternalServerError, "Failed to get product meta fields")
//		return
//	}
//
//	// respond with product meta fields
//	RespondWithJSON(w, http.StatusOK, productMetaFields)
//}
