package handlers

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/mux"
	"net/http"
	"weblineBackend/internal/services"
	"weblineBackend/pkg/utils"
)

type ProductImageHandler struct {
	productImageService *services.ProductService
	s3Client            *s3.Client
	bucketName          string
}

func NewProductImageHandler(productImageService *services.ProductService, s3Client *s3.Client, bucketName string) *ProductImageHandler {
	return &ProductImageHandler{
		productImageService: productImageService,
		s3Client:            s3Client,
		bucketName:          bucketName,
	}
}

// CreateProductImageHandler creates a new product image
func (h *ProductImageHandler) CreateProductImageHandler(w http.ResponseWriter, r *http.Request) {
	filePath, err := utils.UploadFileToS3(r, h.s3Client, h.bucketName, "product-images")
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// get product image ID
	productID := r.FormValue("product_id")

	// create product image
	productImage, err := h.productImageService.CreateProductImage(r.Context(), productID, filePath)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create product image")
		return
	}

	// respond with product image
	RespondWithJSON(w, http.StatusOK, productImage)
}

// GetProductImageByIDHandler retrieves a product image by its ID
func (h *ProductImageHandler) GetProductImageByIDHandler(w http.ResponseWriter, r *http.Request) {
	// get product image ID
	id := mux.Vars(r)["id"]

	// get product image and S3 URL
	productImage, err := h.productImageService.GetProductImageByID(r.Context(), id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get product image")
		return
	}

	// respond with product image and S3 URL
	RespondWithJSON(w, http.StatusOK, productImage)
}

// GetProductImagesByProductIDHandler retrieves all product images by product ID
func (h *ProductImageHandler) GetProductImagesByProductIDHandler(w http.ResponseWriter, r *http.Request) {
	// get product ID
	productID := mux.Vars(r)["product_id"]

	// get product images and S3 URLs
	productImages, err := h.productImageService.ListProductImagesByProductID(r.Context(), productID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get product images")
		return
	}

	// respond with product images and S3 URLs
	RespondWithJSON(w, http.StatusOK, productImages)
}

// UpdateProductImageHandler updates a product image
func (h *ProductImageHandler) UpdateProductImageHandler(w http.ResponseWriter, r *http.Request) {
	filePath, err := utils.UploadFileToS3(r, h.s3Client, h.bucketName, "product-images")
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// get product image ID
	productID := r.FormValue("product_id")

	// update product image
	productImage, err := h.productImageService.UpdateProductImage(r.Context(), productID, filePath)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update product image")
		return
	}

	// respond with product image
	RespondWithJSON(w, http.StatusOK, productImage)
}

// DeleteProductImageHandler deletes a product image
func (h *ProductImageHandler) DeleteProductImageHandler(w http.ResponseWriter, r *http.Request) {
	// Get product image ID from the URL
	id := mux.Vars(r)["id"]

	// Delete product image
	err := h.productImageService.DeleteProductImage(r.Context(), id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete product image")
		return
	}

	// Respond with success
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Product image deleted successfully"})
}
