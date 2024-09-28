package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"weblineBackend/internal/model"
	"weblineBackend/internal/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ProductHandler struct {
	productService          *services.ProductService
	productSEOService       *services.ProductSEOService
	filterService           *services.FilterService
	productAttributeService *services.ProductAttributeService
}

func NewProductHandler(productService *services.ProductService, productSEOService *services.ProductSEOService, filterService *services.FilterService, productAttributeService *services.ProductAttributeService) *ProductHandler {
	return &ProductHandler{
		productService:          productService,
		productSEOService:       productSEOService,
		filterService:           filterService,
		productAttributeService: productAttributeService,
	}
}

// GetFilteredCategoryProducts gets all products by category ID
func (h *ProductHandler) GetFilteredCategoryProducts(w http.ResponseWriter, r *http.Request) {
	// get the filter values
	var body struct {
		AttributeFilters map[string][]string `json:"attributes"`
		MinPrice         float64             `json:"priceFrom"`
		MaxPrice         float64             `json:"priceTo"`
		SortOrder        string              `json:"sortOrder"`
		CategoryID       string              `json:"categoryID"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Println(err)
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// parse category ID
	categoryUUID, err := uuid.Parse(body.CategoryID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	// create params
	params := &model.CategoryProductFilterValues{
		CategoryID:       categoryUUID,
		AttributeFilters: body.AttributeFilters,
		MinPrice:         body.MinPrice,
		MaxPrice:         body.MaxPrice,
		SortOrder:        body.SortOrder,
	}

	// get page and page size
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("limit")

	page, pageSize, err := GetPageAndPageSize(pageStr, pageSizeStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid page or page size: %v", err))
		return
	}
	log.Println("page: ", page, "pageSize: ", pageSize)

	// update the limit and page
	params.Limit = pageSize
	params.Offset = page

	log.Println("params: ", params)
	// get products
	products, err := h.filterService.GetCategoryProductsByFilters(r.Context(), params)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get products by filters")
		return
	}

	RespondWithJSON(w, http.StatusOK, products)
}

// GetAllProductsHandler gets all products
func (h *ProductHandler) GetAllProductsHandler(w http.ResponseWriter, r *http.Request) {
	// get the filter values
	var body struct {
		AttributeFilters map[string][]string `json:"attributes"`
		MinPrice         float64             `json:"priceFrom"`
		MaxPrice         float64             `json:"priceTo"`
		SortOrder        string              `json:"sortOrder"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Println(err)
		RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}

	// create params
	params := &model.AllProductFilterValues{
		AttributeFilters: body.AttributeFilters,
		MinPrice:         body.MinPrice,
		MaxPrice:         body.MaxPrice,
		SortOrder:        body.SortOrder,
	}

	// get page and page size
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("limit")

	page, pageSize, err := GetPageAndPageSize(pageStr, pageSizeStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid page or page size: %v", err))
		return
	}
	log.Println("page: ", page, "pageSize: ", pageSize)

	// update the limit and page
	params.Limit = pageSize
	params.Offset = page

	log.Println("params: ", params)
	// get products
	products, err := h.filterService.GetProductsByFilters(r.Context(), params)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get products by filters")
		return
	}

	RespondWithJSON(w, http.StatusOK, products)
}

// GetProductsByCategoryIDHandler gets all products by category ID
func (h *ProductHandler) GetProductsByCategoryIDHandler(w http.ResponseWriter, r *http.Request) {
	// get category ID
	vars := mux.Vars(r)
	categoryID := vars["id"]

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("limit")

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

// GetFilterOptionsByCategoryNameHandler gets filter options by category name
func (h *ProductHandler) GetFilterOptionsByCategoryNameHandler(w http.ResponseWriter, r *http.Request) {
	// get category name
	categoryName := mux.Vars(r)["name"]

	// get filter options
	filterOptions, err := h.filterService.GetCategoryProductFilterOptions(r.Context(), categoryName)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get filter options")
		return
	}

	// respond with filter options
	RespondWithJSON(w, http.StatusOK, filterOptions)
}

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
	params.MetaTitle = r.FormValue("metaTitle")
	params.MetaDescription = r.FormValue("metaDescription")
	priceStr := r.FormValue("price")
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		log.Println(err)
		RespondWithError(w, http.StatusBadRequest, "Invalid price")
		return
	}
	pricePerUnitStr := r.FormValue("pricePerUnit")
	pricePerUnit, err := strconv.ParseFloat(pricePerUnitStr, 64)
	if err != nil {
		log.Println(err)
		RespondWithError(w, http.StatusBadRequest, "Invalid price per unit")
		return
	}
	params.PricePerUnit = pricePerUnit
	params.Price = price
	stockStr := r.FormValue("stock")
	stock, err := strconv.Atoi(stockStr)
	if err != nil {
		log.Println(err)
		RespondWithError(w, http.StatusBadRequest, "Invalid stock")
		return
	}
	params.Stock = stock
	categoryIDStr := r.FormValue("categoryID")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		log.Println(err)
		RespondWithError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}
	params.CategoryID = categoryID
	params.Status = r.FormValue("status")
	params.PartNumber = r.FormValue("partNumber")
	if params.PartNumber == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid part number")
		return
	}

	// get the images
	images := r.MultipartForm.File["images"]

	// Process image Urls
	imageUrls := make([]string, 0)
	imageUrls = append(imageUrls, r.MultipartForm.Value["url"]...)
	params.ImageUrls = imageUrls

	// Process specifications
	specifications := make([]model.Specification, 0)
	for _, spec := range r.MultipartForm.Value["specifications"] {
		var specification model.Specification
		if err := json.Unmarshal([]byte(spec), &specification); err != nil {
			log.Println(err)
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
		log.Println(err)
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

// GetProductMetaFieldsByCategoryIDHandler gets product meta fields by category ID
func (h *ProductHandler) GetProductMetaFieldsByCategoryIDHandler(w http.ResponseWriter, r *http.Request) {
	// get category ID
	vars := mux.Vars(r)
	categoryID := vars["categoryID"]

	// parse category ID
	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	// get product meta fields
	productMetaFields, err := h.productAttributeService.GetProductAttributesWithValues(r.Context(), categoryUUID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get product meta fields")
		return
	}

	// respond with product meta fields
	RespondWithJSON(w, http.StatusOK, productMetaFields)
}

// GetAllProductFilterOptionsHandler gets all product filter options
func (h *ProductHandler) GetAllProductFilterOptionsHandler(w http.ResponseWriter, r *http.Request) {
	// get filter options
	filterOptions, err := h.filterService.GetProductFilterOptions(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get filter options")
		return
	}

	// respond with filter options
	RespondWithJSON(w, http.StatusOK, filterOptions)
}

// GetProductCartHandler gets a product cart by its slug
func (h *ProductHandler) GetProductCartHandler(w http.ResponseWriter, r *http.Request) {
	// get product slug
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid slug")
		return
	}

	// get product cart
	productCart, err := h.productService.GetProductCart(r.Context(), slug)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get product cart")
		return
	}

	// respond with product cart
	RespondWithJSON(w, http.StatusOK, productCart)
}
