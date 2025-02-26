package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"weblineBackend/internal/app_errors"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/database"
	"weblineBackend/internal/middleware"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
	"weblineBackend/pkg/utils"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	productsCacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_cache_hits_total",
		Help: "The total number of cache hits for products",
	})
	productsCacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_cache_misses_total",
		Help: "The total number of cache misses for products",
	})
	productsRetrievalTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "products_retrieval_duration_seconds",
		Help:    "The duration of products retrieval in seconds",
		Buckets: prometheus.DefBuckets,
	})
	productCreationDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "product_creation_duration_seconds",
		Help: "Duration of product creation in seconds",
	})
	productCreationTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "product_creation_total",
		Help: "Total number of product creations",
	})
	productCreationErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "product_creation_errors_total",
		Help: "Total number of product creation errors",
	})
)

type ProductService struct {
	productRepo              *repository.ProductRepository
	productVariantRepo       *repository.ProductVariantRepository
	productImageRepo         *repository.ProductImageRepository
	productSpecificationRepo *repository.ProductSpecificationRepository
	categoryRepo             *repository.CategoryRepository
	productOptionRepo        *repository.ProductOptionRepository
	discountRepo             *repository.DiscountRepository
	userRepo                 *repository.UserRepository
	exchangeRateRepo         repository.ExchangeRateRepository
	cacheService             CacheService
	logger                   *zap.Logger
	config                   *appconfig.Config
	s3Client                 *s3.Client
}

func NewProductService(
	productRepo *repository.ProductRepository,
	productVariantRepo *repository.ProductVariantRepository,
	productImageRepo *repository.ProductImageRepository,
	productSpecificationRepo *repository.ProductSpecificationRepository,
	categoryRepo *repository.CategoryRepository,
	productOptionRepo *repository.ProductOptionRepository,
	discountRepo *repository.DiscountRepository,
	userRepo *repository.UserRepository,
	exchangeRateRepo repository.ExchangeRateRepository,
	cacheService CacheService,
	logger *zap.Logger,
	config *appconfig.Config,
	s3Client *s3.Client,
) *ProductService {
	return &ProductService{
		productRepo:              productRepo,
		productVariantRepo:       productVariantRepo,
		productImageRepo:         productImageRepo,
		productSpecificationRepo: productSpecificationRepo,
		categoryRepo:             categoryRepo,
		productOptionRepo:        productOptionRepo,
		discountRepo:             discountRepo,
		userRepo:                 userRepo,
		exchangeRateRepo:         exchangeRateRepo,
		cacheService:             cacheService,
		logger:                   logger,
		config:                   config,
		s3Client:                 s3Client,
	}
}

// GetProductBySlug retrieves a product by its slug
func (s *ProductService) GetProductBySlug(ctx context.Context, slug string) (model.ProductDetail, error) {
	// Use standardized cache key function
	cacheKey := ProductDetailKey(slug)
	var product model.ProductDetail

	// Try to get product from CacheService
	err := s.cacheService.Get(ctx, cacheKey, &product)
	if err == nil && product.Name != "" {
		s.logger.Debug("Product retrieved from cache", zap.String("slug", slug))
		return product, nil
	}

	// If not in cache or error occurred, fetch from database
	dbProduct, err := s.productRepo.GetProductBySlug(ctx, slug)
	if err != nil {
		s.logger.Error("Failed to get product by slug", zap.String("slug", slug), zap.Error(err))
		return model.ProductDetail{}, fmt.Errorf("failed to get product by slug: %w", err)
	}

	// Map database product to model
	product = model.ProductDetail{
		Name: dbProduct.Name,
		Slug: dbProduct.Slug,
	}

	// Cache the product for future requests
	err = s.cacheService.Set(ctx, cacheKey, product)
	if err != nil {
		s.logger.Warn("Failed to cache product", zap.String("key", cacheKey), zap.Error(err))
		// Proceed without failing if caching fails
	}

	return product, nil
}

func (s *ProductService) calculatePriceToKES(ctx context.Context, price string) (float64, error) {
	// Parse the price
	p, err := strconv.ParseFloat(price, 64)
	if err != nil {
		s.logger.Error("failed to parse product price", zap.Error(err))
		return 0, fmt.Errorf("failed to parse product price: %w", err)
	}

	// Get the exchange rate
	exchangeRate, err := s.exchangeRateRepo.GetLatestExchangeRate(ctx, "USD")
	if err != nil {
		s.logger.Error("failed to get exchange rate by currency", zap.Error(err))
		return 0, fmt.Errorf("failed to get exchange rate by currency: %w", err)
	}

	// Calculate the price in KES
	priceToKES := p * exchangeRate

	// Apply rounding
	roundedPrice := utils.RoundPrice(priceToKES)

	return roundedPrice, nil
}

func (s *ProductService) getProductDiscountPercentage(ctx context.Context, productID uuid.UUID) (float64, error) {
	discount, err := s.discountRepo.GetDiscountByProductID(ctx, &productID)
	if err != nil {
		s.logger.Error("failed to get product discount", zap.Error(err))
		return 0, fmt.Errorf("failed to get product discount: %w", err)
	}

	if discount != nil {
		discountPercentage, err := strconv.ParseFloat(discount.DiscountPercentage, 64)
		if err != nil {
			s.logger.Error("failed to parse discount percentage", zap.Error(err))
			return 0, fmt.Errorf("failed to parse discount percentage: %w", err)
		}
		return discountPercentage, nil
	}

	return 0, nil
}

// Adjusted method to use standardized cache keys and improved error handling
func (s *ProductService) GetProductsByCategoryID(ctx context.Context, categoryID string, pageSize int32, page int32) (model.PaginationResult[[]model.Product], error) {
	// Parse the category ID
	categoryIDValue, err := uuid.Parse(categoryID)
	if err != nil {
		s.logger.Error("Invalid category ID format", zap.String("categoryID", categoryID), zap.Error(err))
		return model.PaginationResult[[]model.Product]{}, fmt.Errorf("invalid category ID format: %w", err)
	}

	// Create a cache key using standardized function
	cacheKey := GenerateCacheKey(NamespaceProduct, "category", categoryID, "page", strconv.Itoa(int(page)), "size", strconv.Itoa(int(pageSize)))

	var result model.PaginationResult[[]model.Product]
	// Try to get the result from cache
	err = s.cacheService.Get(ctx, cacheKey, &result)
	if err == nil && len(result.Data) > 0 {
		s.logger.Debug("Products by category retrieved from cache", zap.String("categoryID", categoryID))
		return result, nil
	}

	// If not in cache, proceed with database query
	totalProductsByCategory, err := s.productRepo.CountProductsByParentCategoryID(ctx, categoryIDValue)
	if err != nil {
		s.logger.Error("Failed to count products by category ID", zap.Error(err))
		return model.PaginationResult[[]model.Product]{}, fmt.Errorf("failed to count products by category ID: %w", err)
	}

	if totalProductsByCategory == 0 {
		return model.PaginationResult[[]model.Product]{}, nil
	}

	paginatedProductsByCategory, err := utils.Paginate(
		s.config,
		totalProductsByCategory,
		page,
		pageSize,
		func(offset int32, limit int32) ([]model.Product, error) {
			products, err := s.productRepo.GetProductsByCategoryID(ctx, categoryIDValue, limit, offset)
			if err != nil {
				s.logger.Error("Failed to get products by category ID", zap.Error(err))
				return nil, fmt.Errorf("failed to get products by category ID: %w", err)
			}

			return s.mapProductsToModel(ctx, products)
		},
	)

	if err != nil {
		s.logger.Error("Failed to paginate products by category ID", zap.Error(err))
		return model.PaginationResult[[]model.Product]{}, fmt.Errorf("failed to paginate products by category ID: %w", err)
	}

	// Cache the result
	err = s.cacheService.Set(ctx, cacheKey, *paginatedProductsByCategory)
	if err != nil {
		s.logger.Warn("Failed to cache products by category ID", zap.Error(err))
		// Continue without failing if caching fails
	}

	s.logger.Debug("Products by category retrieved from database and cached", zap.String("categoryID", categoryID))
	return *paginatedProductsByCategory, nil
}

// mapProductsToModel maps database products to model products with rounded prices
func (s *ProductService) mapProductsToModel(ctx context.Context, products []model.ProductSchema) ([]model.Product, error) {
	var productSchemas []model.Product
	for _, product := range products {
		productImages, err := s.productImageRepo.ListProductImagesByProductID(ctx, uuid.NullUUID{UUID: product.ID, Valid: true})
		if err != nil {
			s.logger.Error("failed to get product images", zap.Error(err))
			return nil, fmt.Errorf("failed to get product images: %w", err)
		}

		// Calculate price to KES
		price, err := s.calculatePriceToKES(ctx, product.USD)
		if err != nil {
			s.logger.Error("failed to calculate price to KES", zap.Error(err))
			return nil, err
		}

		finalPrice := utils.RoundPrice(price)

		discountPercentage, err := s.getProductDiscountPercentage(ctx, product.ID)
		if err != nil {
			return nil, err
		}

		imageUrl := ""
		if len(productImages) > 0 {
			imageUrl = s.constructS3URL(productImages[0].ImageUrl)
		}

		// Use rounded price
		productSchemas = append(productSchemas, model.Product{
			ID:              product.ID,
			Name:            product.Name,
			Description:     product.Description,
			Price:           fmt.Sprintf("%.2f", finalPrice), // Rounded price
			Slug:            product.Slug,
			ImageURL:        imageUrl,
			DiscountPercent: discountPercentage,
		})
	}
	return productSchemas, nil
}

// SearchProducts searches for products by name or description
func (s *ProductService) SearchProducts(ctx context.Context, searchTerm string) ([]model.ProductQueryResult, error) {
	// Fetch products by search term from the repository
	products, err := s.productRepo.SearchProducts(ctx, searchTerm)
	if err != nil {
		s.logger.Error("failed to search products", zap.Error(err))
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	// Map the products to the result model
	return s.mapProductsToQueryResult(ctx, products)
}

func (s *ProductService) mapProductsToQueryResult(ctx context.Context, products []model.ProductSchema) ([]model.ProductQueryResult, error) {
	var productsQueryResult []model.ProductQueryResult

	for _, product := range products {
		// Get the primary image URL for the product
		imageURL, err := s.getPrimaryProductImageURL(ctx, product.ID)
		if err != nil {
			return nil, err
		}

		// Get the discount percentage for the product
		discountPercentage, err := s.getProductDiscountPercentage(ctx, product.ID)
		if err != nil {
			return nil, err
		}

		// Calculate price to KES
		price, err := s.calculatePriceToKES(ctx, product.USD)
		if err != nil {
			s.logger.Error("failed to calculate price to KES", zap.Error(err))
			return nil, err
		}

		finalPrice := utils.RoundPrice(price)

		productsQueryResult = append(productsQueryResult, model.ProductQueryResult{
			ID:              product.ID,
			Name:            product.Name,
			Price:           fmt.Sprintf("%.2f", finalPrice),
			Stock:           product.Stock,
			ImageURL:        imageURL,
			DiscountPercent: discountPercentage,
			Slug:            product.Slug,
		})
	}

	return productsQueryResult, nil
}

func (s *ProductService) getPrimaryProductImageURL(ctx context.Context, productID uuid.UUID) (string, error) {
	productImages, err := s.productImageRepo.ListProductImagesByProductID(ctx, uuid.NullUUID{UUID: productID, Valid: true})
	if err != nil {
		s.logger.Error("failed to get product images", zap.Error(err))
		return "", fmt.Errorf("failed to get product images: %w", err)
	}

	if len(productImages) == 0 {
		return "", nil
	}

	return s.constructS3URL(productImages[0].ImageUrl), nil
}

// CreateProductVariant creates a new product variant
func (s *ProductService) CreateProductVariant(
	ctx context.Context,
	productID string,
	variantName string,
	variantValue string,
	price float64,
	stock int32,
) (database.ProductVariant, error) {
	// Parsing
	var productIDValue uuid.NullUUID
	if productID != "" {
		productIDValue.UUID, _ = uuid.Parse(productID)
		productIDValue.Valid = true
	} else {
		productIDValue.Valid = false
	}

	params := database.CreateProductVariantParams{
		ProductID:    productIDValue,
		VariantName:  variantName,
		VariantValue: variantValue,
		Price:        strconv.FormatFloat(price, 'f', 2, 64),
		Stock:        sql.NullInt32{Int32: stock, Valid: true},
	}

	// Create product variant
	productVariant, err := s.productVariantRepo.CreateProductVariant(ctx, params)
	if err != nil {
		s.logger.Error("failed to create product variant", zap.Error(err))
		return database.ProductVariant{}, err
	}

	return productVariant, nil
}

// GetProductVariantByID retrieves a product variant by its ID
func (s *ProductService) GetProductVariantByID(ctx context.Context, variantID string) (database.ProductVariant, error) {
	// Parse the variant ID
	id, err := uuid.Parse(variantID)
	if err != nil {
		s.logger.Error("invalid product variant ID format", zap.String("variantID", variantID), zap.Error(err))
		return database.ProductVariant{}, fmt.Errorf("invalid product variant ID format: %w", err)
	}

	// Get the product variant from the repository
	productVariant, err := s.productVariantRepo.GetProductVariantByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get product variant by ID", zap.String("variantID", variantID), zap.Error(err))
		return database.ProductVariant{}, fmt.Errorf("failed to get product variant by ID: %w", err)
	}

	return productVariant, nil
}

// ListProductVariantsByProductID retrieves all product variants by their product ID
func (s *ProductService) ListProductVariantsByProductID(ctx context.Context, productID string) ([]database.ProductVariant, error) {
	// Parse the product ID
	id, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("invalid product ID format", zap.String("productID", productID), zap.Error(err))
		return nil, fmt.Errorf("invalid product ID format: %w", err)
	}

	// Get the product variants from the repository
	productVariants, err := s.productVariantRepo.ListProductVariantsByProductID(ctx, uuid.NullUUID{UUID: id, Valid: true})
	if err != nil {
		s.logger.Error("failed to list product variants by product ID", zap.Error(err))
		return nil, fmt.Errorf("failed to list product variants by product ID: %w", err)
	}

	return productVariants, nil
}

// UpdateProductVariant updates an existing product variant
func (s *ProductService) UpdateProductVariant(
	ctx context.Context,
	id string,
	variantName string,
	variantValue string,
	price float64,
	stock int32,
) (database.ProductVariant, error) {
	// Parse the product variant ID
	productVariantID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid product variant ID", zap.Error(err))
		return database.ProductVariant{}, fmt.Errorf("invalid product variant ID: %w", err)
	}

	// Prepare the parameters for the update
	params := database.UpdateProductVariantParams{
		ID:           productVariantID,
		VariantName:  variantName,
		VariantValue: variantValue,
		Price:        strconv.FormatFloat(price, 'f', 2, 64),
		Stock:        sql.NullInt32{Int32: stock, Valid: true},
	}

	// Update the product variant in the repository
	productVariant, err := s.productVariantRepo.UpdateProductVariant(ctx, params)
	if err != nil {
		s.logger.Error("failed to update product variant", zap.Error(err))
		return database.ProductVariant{}, fmt.Errorf("failed to update product variant: %w", err)
	}

	return productVariant, nil
}

// DeleteProductVariant deletes a product variant from the database
func (s *ProductService) DeleteProductVariant(ctx context.Context, id string) error {
	// Parse the product variant ID
	productVariantID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid product variant ID", zap.Error(err))
		return fmt.Errorf("invalid product variant ID: %w", err)
	}

	// Delete the product variant from the repository
	err = s.productVariantRepo.DeleteProductVariant(ctx, productVariantID)
	if err != nil {
		s.logger.Error("failed to delete product variant", zap.Error(err))
		return fmt.Errorf("failed to delete product variant: %w", err)
	}

	return nil
}

// CreateProductImage creates a new product image
func (s *ProductService) CreateProductImage(
	ctx context.Context,
	r *http.Request,
	productID string,
) (model.ProductImage, error) {
	// Parse the product ID
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("invalid product ID", zap.Error(err))
		return model.ProductImage{}, fmt.Errorf("invalid product ID: %w", err)
	}

	// Check if the product exists
	if _, err := s.productRepo.GetProductByID(ctx, productUUID); err != nil {
		s.logger.Error("product does not exist", zap.Error(err))
		return model.ProductImage{}, fmt.Errorf("product does not exist: %w", err)
	}

	// Upload the image to the s3 bucket
	filePath, err := utils.UploadFileToS3(ctx, r, s.s3Client, s.config.AWSBucketName, "product-images")
	if err != nil {
		s.logger.Error("failed to upload image to s3", zap.Error(err))
		return model.ProductImage{}, fmt.Errorf("failed to upload image to s3: %w", err)
	}

	// Prepare the parameters for creating the product image
	params := database.CreateProductImageParams{
		ProductID: uuid.NullUUID{UUID: productUUID, Valid: true},
		ImageUrl:  filePath,
	}

	// Create the product image in the repository
	productImage, err := s.productImageRepo.CreateProductImage(ctx, params)
	if err != nil {
		s.logger.Error("failed to create product image", zap.Error(err))
		return model.ProductImage{}, fmt.Errorf("failed to create product image: %w", err)
	}

	return s.mapToProductImageModel(productImage), nil
}

// GetProductImageByID retrieves a product image by its ID
func (s *ProductService) GetProductImageByID(ctx context.Context, imageID string) (model.ProductImage, error) {
	// Parse the image ID
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		s.logger.Error("invalid product image ID", zap.Error(err))
		return model.ProductImage{}, fmt.Errorf("invalid product image ID: %w", err)
	}

	// Get the product image from the repository
	productImage, err := s.productImageRepo.GetProductImageByID(ctx, imageUUID)
	if err != nil {
		s.logger.Error("failed to get product image by ID", zap.Error(err))
		return model.ProductImage{}, fmt.Errorf("failed to get product image by ID: %w", err)
	}

	return s.mapToProductImageModel(productImage), nil
}

// ListProductImagesByProductID retrieves all product images by their product ID
func (s *ProductService) ListProductImagesByProductID(ctx context.Context, productID string) ([]model.ProductImage, error) {
	// Parse the product ID
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("invalid product ID", zap.Error(err))
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	// Get the product images from the repository
	productImages, err := s.productImageRepo.ListProductImagesByProductID(ctx, uuid.NullUUID{UUID: productUUID, Valid: true})
	if err != nil {
		s.logger.Error("failed to list product images by product ID", zap.Error(err))
		return nil, fmt.Errorf("failed to list product images by product ID: %w", err)
	}

	return s.mapToProductImageModels(productImages), nil
}

// UpdateProductImage updates an existing product image
func (s *ProductService) UpdateProductImage(
	ctx context.Context,
	r *http.Request,
	id string,
) (model.ProductImage, error) {
	// Parse the image ID
	imageUUID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid product image ID", zap.Error(err))
		return model.ProductImage{}, fmt.Errorf("invalid product image ID: %w", err)
	}

	// Check if the product image exists
	if _, err := s.GetProductImageByID(ctx, id); err != nil {
		s.logger.Error("product image does not exist", zap.Error(err))
		return model.ProductImage{}, fmt.Errorf("product image does not exist: %w", err)
	}

	// Upload the image to the s3 bucket
	filePath, err := utils.UploadFileToS3(ctx, r, s.s3Client, s.config.AWSBucketName, "product-images")
	if err != nil {
		s.logger.Error("failed to upload image to s3", zap.Error(err))
		return model.ProductImage{}, fmt.Errorf("failed to upload image to s3: %w", err)
	}

	// Prepare the parameters for updating the product image
	params := database.UpdateProductImageParams{
		ID:       imageUUID,
		ImageUrl: filePath,
	}

	// Update the product image in the repository
	productImage, err := s.productImageRepo.UpdateProductImage(ctx, params)
	if err != nil {
		s.logger.Error("failed to update product image", zap.Error(err))
		return model.ProductImage{}, fmt.Errorf("failed to update product image: %w", err)
	}

	return s.mapToProductImageModel(productImage), nil
}

// mapToProductImageModel maps a database ProductImage to a model ProductImage
func (s *ProductService) mapToProductImageModel(dbImage database.ProductImage) model.ProductImage {
	return model.ProductImage{
		ID:        dbImage.ID,
		ProductID: dbImage.ProductID.UUID.String(),
		S3URL:     s.constructS3URL(dbImage.ImageUrl),
	}
}

// mapToProductImageModels maps a slice of database ProductImage to a slice of model ProductImage
func (s *ProductService) mapToProductImageModels(dbImages []database.ProductImage) []model.ProductImage {
	images := make([]model.ProductImage, len(dbImages))
	for i, dbImage := range dbImages {
		images[i] = s.mapToProductImageModel(dbImage)
	}
	return images
}

// constructS3URL constructs the S3 URL for a given file path
func (s *ProductService) constructS3URL(filePath string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.AWSBucketName, s.config.AWSRegion, filePath)
}

// getFilePathFromS3URL extracts the file path from an S3 URL
func (s *ProductService) getFilePathFromS3URL(s3URL string) string {
	parsedURL, err := url.Parse(s3URL)
	if err != nil {
		s.logger.Error("failed to parse S3 URL", zap.Error(err))
		return ""
	}

	// Remove the bucket name and leading slash
	path := strings.TrimPrefix(parsedURL.Path, fmt.Sprintf("/%s/", s.config.AWSBucketName))
	// Ensure no leading slash
	path = strings.TrimPrefix(path, "/")

	return path
}

// DeleteProductImage deletes a product image from the database and S3
func (s *ProductService) DeleteProductImage(ctx context.Context, id string) error {
	// Parse the image ID
	imageUUID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid product image ID", zap.Error(err))
		return fmt.Errorf("invalid product image ID: %w", err)
	}

	// Fetch the product image details to get the S3 key (file path)
	productImage, err := s.productImageRepo.GetProductImageByID(ctx, imageUUID)
	if err != nil {
		s.logger.Error("failed to get product image by ID", zap.Error(err))
		return fmt.Errorf("failed to get product image by ID: %w", err)
	}

	// Delete the product image from the repository
	err = s.productImageRepo.DeleteProductImage(ctx, imageUUID)
	if err != nil {
		s.logger.Error("failed to delete product image from database", zap.Error(err))
		return fmt.Errorf("failed to delete product image from database: %w", err)
	}

	// Extract S3 key from the URL
	s3Key := strings.TrimPrefix(productImage.ImageUrl, fmt.Sprintf("https://%s.s3.%s.amazonaws.com/", s.config.AWSBucketName, s.config.AWSRegion))

	// Delete the product image from S3
	_, err = s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.config.AWSBucketName,
		Key:    &s3Key,
	})
	if err != nil {
		s.logger.Error("failed to delete product image from S3", zap.Error(err))
		return fmt.Errorf("failed to delete product image from S3: %w", err)
	}

	return nil
}

// CreateProductSpecification creates a new product specification
func (s *ProductService) CreateProductSpecification(
	ctx context.Context,
	productID string,
	specName string,
	specValue string,
) (database.ProductSpecification, error) {
	// Parse the product ID
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("invalid product ID", zap.Error(err))
		return database.ProductSpecification{}, fmt.Errorf("invalid product ID: %w", err)
	}

	// Prepare parameters for creating product specification
	params := database.CreateProductSpecificationParams{
		ProductID: uuid.NullUUID{UUID: productUUID, Valid: true},
		SpecName:  specName,
		SpecValue: specValue,
	}

	// Create product specification
	productSpec, err := s.productRepo.CreateProductSpecification(ctx, params)
	if err != nil {
		s.logger.Error("failed to create product specification", zap.Error(err))
		return database.ProductSpecification{}, fmt.Errorf("failed to create product specification: %w", err)
	}

	return productSpec, nil
}

// ListProductSpecificationsByProductID lists product specifications by product ID
func (s *ProductService) ListProductSpecificationsByProductID(
	ctx context.Context,
	productID string,
) ([]database.ProductSpecification, error) {
	// Parse the product ID
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("invalid product ID", zap.Error(err))
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	// List product specifications by product ID
	productSpecs, err := s.productRepo.ListProductSpecificationsByProductID(ctx, uuid.NullUUID{UUID: productUUID, Valid: true})
	if err != nil {
		s.logger.Error("failed to list product specifications", zap.Error(err))
		return nil, fmt.Errorf("failed to list product specifications: %w", err)
	}

	return productSpecs, nil
}

// UpdateProductSpecification updates an existing product specification
func (s *ProductService) UpdateProductSpecification(
	ctx context.Context,
	id string,
	specName string,
	specValue string,
) (database.ProductSpecification, error) {
	// Parse the product specification ID
	specUUID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid product specification ID", zap.Error(err))
		return database.ProductSpecification{}, fmt.Errorf("invalid product specification ID: %w", err)
	}

	// Prepare parameters for updating product specification
	params := database.UpdateProductSpecificationParams{
		ID:        specUUID,
		SpecName:  specName,
		SpecValue: specValue,
	}

	// Update product specification
	productSpec, err := s.productRepo.UpdateProductSpecification(ctx, params)
	if err != nil {
		s.logger.Error("failed to update product specification", zap.Error(err))
		return database.ProductSpecification{}, fmt.Errorf("failed to update product specification: %w", err)
	}

	return productSpec, nil
}

// DeleteProductSpecification deletes a product specification from the database
func (s *ProductService) DeleteProductSpecification(ctx context.Context, id string) error {
	// Parse the product specification ID
	specUUID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid product specification ID", zap.Error(err))
		return fmt.Errorf("invalid product specification ID: %w", err)
	}

	// Delete product specification
	err = s.productRepo.DeleteProductSpecification(ctx, specUUID)
	if err != nil {
		s.logger.Error("failed to delete product specification", zap.Error(err))
		return fmt.Errorf("failed to delete product specification: %w", err)
	}

	return nil
}

// CreateProductOption creates a new product option
func (s *ProductService) CreateProductOption(
	ctx context.Context,
	productID string,
	optionName string,
) (database.ProductOption, error) {
	// Parse the product ID
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("invalid product ID", zap.Error(err))
		return database.ProductOption{}, fmt.Errorf("invalid product ID: %w", err)
	}

	// Prepare parameters for creating product option
	params := database.CreateProductOptionParams{
		ProductID:  uuid.NullUUID{UUID: productUUID, Valid: true},
		OptionName: optionName,
	}

	// Create product option
	productOption, err := s.productOptionRepo.CreateProductOption(ctx, params)
	if err != nil {
		s.logger.Error("failed to create product option", zap.Error(err))
		return database.ProductOption{}, fmt.Errorf("failed to create product option: %w", err)
	}

	return productOption, nil
}

// ListProductOptionsByProductID lists product options by product ID
func (s *ProductService) ListProductOptionsByProductID(
	ctx context.Context,
	productID string,
) ([]database.ProductOption, error) {
	// Parse the product ID
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("invalid product ID", zap.Error(err))
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	// List product options by product ID
	productOptions, err := s.productOptionRepo.ListProductOptionsByProductID(ctx, uuid.NullUUID{UUID: productUUID, Valid: true})
	if err != nil {
		s.logger.Error("failed to list product options", zap.Error(err))
		return nil, fmt.Errorf("failed to list product options: %w", err)
	}

	return productOptions, nil
}

// UpdateProductOption updates an existing product option
func (s *ProductService) UpdateProductOption(
	ctx context.Context,
	id string,
	optionName string,
) (database.ProductOption, error) {
	// Parse the product option ID
	optionUUID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid product option ID", zap.Error(err))
		return database.ProductOption{}, fmt.Errorf("invalid product option ID: %w", err)
	}

	// Prepare parameters for updating product option
	params := database.UpdateProductOptionParams{
		ID:         optionUUID,
		OptionName: optionName,
	}

	// Update product option
	productOption, err := s.productOptionRepo.UpdateProductOption(ctx, params)
	if err != nil {
		s.logger.Error("failed to update product option", zap.Error(err))
		return database.ProductOption{}, fmt.Errorf("failed to update product option: %w", err)
	}

	return productOption, nil
}

// DeleteProductOption deletes a product option from the database
func (s *ProductService) DeleteProductOption(ctx context.Context, id string) error {
	// Parse the product option ID
	optionUUID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid product option ID", zap.Error(err))
		return fmt.Errorf("invalid product option ID: %w", err)
	}

	// Delete product option
	err = s.productOptionRepo.DeleteProductOption(ctx, optionUUID)
	if err != nil {
		s.logger.Error("failed to delete product option", zap.Error(err))
		return fmt.Errorf("failed to delete product option: %w", err)
	}

	return nil
}

// CreateProductOptionValue creates a new product option value
func (s *ProductService) CreateProductOptionValue(
	ctx context.Context,
	optionID string,
	optionValue string,
	additionalPrice float64,
) (database.ProductOptionValue, error) {
	// Parse the option ID
	optionUUID, err := uuid.Parse(optionID)
	if err != nil {
		s.logger.Error("invalid product option ID", zap.Error(err))
		return database.ProductOptionValue{}, fmt.Errorf("invalid product option ID: %w", err)
	}

	// Prepare parameters for creating product option value
	params := database.CreateProductOptionValueParams{
		OptionID:        uuid.NullUUID{UUID: optionUUID, Valid: true},
		ValueName:       optionValue,
		AdditionalPrice: sql.NullString{String: strconv.FormatFloat(additionalPrice, 'f', 2, 64), Valid: true},
	}

	// Create product option value
	productOptionValue, err := s.productOptionRepo.CreateProductOptionValue(ctx, params)
	if err != nil {
		s.logger.Error("failed to create product option value", zap.Error(err))
		return database.ProductOptionValue{}, fmt.Errorf("failed to create product option value: %w", err)
	}

	return productOptionValue, nil
}

// ListProductOptionValuesByOptionID lists product option values by option ID
func (s *ProductService) ListProductOptionValuesByOptionID(
	ctx context.Context,
	optionID string,
) ([]database.ProductOptionValue, error) {
	// Parse the option ID
	optionUUID, err := uuid.Parse(optionID)
	if err != nil {
		s.logger.Error("invalid product option ID", zap.Error(err))
		return nil, fmt.Errorf("invalid product option ID: %w", err)
	}

	// List product option values by option ID
	productOptionValues, err := s.productOptionRepo.ListProductOptionValuesByOptionID(ctx, uuid.NullUUID{UUID: optionUUID, Valid: true})
	if err != nil {
		s.logger.Error("failed to list product option values", zap.Error(err))
		return nil, fmt.Errorf("failed to list product option values: %w", err)
	}

	return productOptionValues, nil
}

// UpdateProductOptionValue updates an existing product option value
func (s *ProductService) UpdateProductOptionValue(
	ctx context.Context,
	id string,
	optionValue string,
	additionalPrice float64,
) (database.ProductOptionValue, error) {
	// Parse the option value ID
	optionValueUUID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid product option value ID", zap.Error(err))
		return database.ProductOptionValue{}, fmt.Errorf("invalid product option value ID: %w", err)
	}

	// Prepare parameters for updating product option value
	params := database.UpdateProductOptionValueParams{
		ID:              optionValueUUID,
		ValueName:       optionValue,
		AdditionalPrice: sql.NullString{String: strconv.FormatFloat(additionalPrice, 'f', 2, 64), Valid: true},
	}

	// Update product option value
	productOptionValue, err := s.productOptionRepo.UpdateProductOptionValue(ctx, params)
	if err != nil {
		s.logger.Error("failed to update product option value", zap.Error(err))
		return database.ProductOptionValue{}, fmt.Errorf("failed to update product option value: %w", err)
	}

	return productOptionValue, nil
}

// DeleteProductOptionValue deletes a product option value from the database
func (s *ProductService) DeleteProductOptionValue(ctx context.Context, id string) error {
	// Parse the option value ID
	optionValueUUID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid product option value ID", zap.Error(err))
		return fmt.Errorf("invalid product option value ID: %w", err)
	}

	// Delete product option value
	err = s.productOptionRepo.DeleteProductOptionValue(ctx, optionValueUUID)
	if err != nil {
		s.logger.Error("failed to delete product option value", zap.Error(err))
		return fmt.Errorf("failed to delete product option value: %w", err)
	}

	return nil
}

type FilterOptions struct {
	Categories    []model.ProductCategoryFilterOption `json:"categories"`
	TotalProducts int64                               `json:"totalProducts"`
	Processor     []string                            `json:"processor"`
	Storage       []string                            `json:"storage"`
	Color         []string                            `json:"color"`
	Size          []string                            `json:"size"`
}

// GetAllProductSitemap retrieves all products for sitemap, using cache when available
func (s *ProductService) GetAllProductSitemap(ctx context.Context) ([]*model.ProductSitemap, error) {
	cacheKey := ProductSitemapKey()
	var productSitemap []*model.ProductSitemap

	err := s.cacheService.GetOrSet(ctx, cacheKey, &productSitemap, func() error {
		// Fetch all products from the repository
		products, err := s.productRepo.ListProducts(ctx, database.ListProductsParams{
			Offset: 0,
			Limit:  100, // Adjust limit as per your requirements
		})
		if err != nil {
			s.logger.Error("Failed to get all products for sitemap", zap.Error(err))
			return fmt.Errorf("failed to get all products for sitemap: %w", err)
		}

		// Transform products into ProductSitemap
		productSitemap = make([]*model.ProductSitemap, 0, len(products))
		for _, product := range products {
			productSitemap = append(productSitemap, &model.ProductSitemap{
				ID:        product.ID,
				UpdatedAt: product.UpdatedAt.Time,
			})
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to get product sitemap", zap.Error(err))
		return nil, err
	}

	s.logger.Debug("Product sitemap retrieved from cache or database")
	return productSitemap, nil
}

// Updated GetProducts retrieves paginated products with caching
func (s *ProductService) GetProducts(ctx context.Context, page int32, pageSize int32) (*model.PaginationResult[[]*model.V2Product], error) {
	timer := prometheus.NewTimer(productsRetrievalTime)
	defer timer.ObserveDuration()

	// Ensure valid pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10 // or use a default from your configuration
	}

	// Use a cache key that reflects the pagination parameters
	cacheKey := ProductAllPaginatedKey(page, pageSize)
	var paginatedResult model.PaginationResult[[]*model.V2Product]

	// Try to get paginated products from cache
	err := s.cacheService.Get(ctx, cacheKey, &paginatedResult)
	if err == nil && paginatedResult.Data != nil && len(paginatedResult.Data) > 0 {
		productsCacheHits.Inc()
		s.logger.Debug("Paginated products retrieved from cache")
		return &paginatedResult, nil
	}
	productsCacheMisses.Inc()

	// Get total count of products for pagination metadata
	totalCount, err := s.productRepo.CountProducts(ctx)
	if err != nil {
		s.logger.Error("Failed to count products", zap.Error(err))
		return nil, err
	}

	// Fetch paginated products from the repository (using page and pageSize)
	products, err := s.productRepo.GetV2Products(ctx, page, pageSize)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Error("No products found")
			return nil, err
		}
		s.logger.Error("Failed to get products", zap.Error(err))
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	// Post-process each product (update price and image URL)
	for _, product := range products {
		product.Price = utils.RoundPriceString(product.Price)
		if product.ImageURL != "" {
			product.ImageURL = s.constructS3URL(product.ImageURL)
		}
	}

	// Calculate total number of pages
	totalPages := (totalCount + int64(pageSize) - 1) / int64(pageSize)
	paginatedResult = model.PaginationResult[[]*model.V2Product]{
		TotalCount: totalCount,
		TotalPages: int32(totalPages),
		Page:       int32(page),
		PageSize:   int32(pageSize),
		Data:       products,
	}

	// Cache the paginated result
	if err := s.cacheService.Set(ctx, cacheKey, paginatedResult); err != nil {
		s.logger.Warn("Failed to cache paginated products", zap.Error(err))
	}

	s.logger.Debug("Paginated products retrieved from database and cached")
	return &paginatedResult, nil
}

// GetProductDetail retrieves a product by slug, using cache when available
func (s *ProductService) GetProductDetail(ctx context.Context, slug string) (*model.V2ProductDetail, error) {
	cacheKey := ProductDetailKey(slug)
	var product model.V2ProductDetail

	err := s.cacheService.GetOrSet(ctx, cacheKey, &product, func() error {
		fetchedProduct, err := s.productRepo.GetV2ProductDetailBySlug(ctx, slug)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.logger.Error("Product not found", zap.String("slug", slug))
				return fmt.Errorf("product not found: %w", err)
			}
			s.logger.Error("Failed to get product", zap.Error(err))
			return fmt.Errorf("failed to get product: %w", err)
		}

		product = *fetchedProduct

		var images []model.V2ProductImage
		if err := json.Unmarshal(product.Images, &images); err != nil {
			s.logger.Error("Failed to unmarshal images", zap.Error(err))
			return fmt.Errorf("failed to unmarshal images: %w", err)
		}

		// Update the product image URLs
		for i := range images {
			if images[i].Url != "" {
				images[i].Url = s.constructS3URL(images[i].Url)
			}
		}

		updatedImages, err := json.Marshal(images)
		if err != nil {
			s.logger.Error("Failed to marshal images", zap.Error(err))
			return fmt.Errorf("failed to marshal images: %w", err)
		}

		product.Images = updatedImages
		return nil
	})
	if err != nil {
		s.logger.Error("Failed to get product detail", zap.Error(err))
		return nil, err
	}

	s.logger.Debug("Product detail retrieved from cache or database", zap.String("slug", slug))
	return &product, nil
}

func (s *ProductService) getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return uuid.Nil, app_errors.NewUnauthorizedUserError()
	}
	return userID, nil
}

func (s *ProductService) verifyAdminStatus(ctx context.Context, userID uuid.UUID) error {
	isAdmin, err := s.userRepo.IsAdmin(ctx, userID)
	if err != nil || !isAdmin {
		return app_errors.NewUnauthorizedUserError()
	}
	return nil
}

// NEW HELPER FUNCTION: Invalidate all product-related caches
func (s *ProductService) invalidateProductCache(ctx context.Context) {
	pattern := ProductCachePattern() // e.g. "product:*"
	if err := s.cacheService.DeleteKeysByPattern(ctx, pattern); err != nil {
		s.logger.Warn("failed to invalidate product cache by pattern", zap.String("pattern", pattern), zap.Error(err))
	} else {
		s.logger.Debug("product cache invalidated using pattern", zap.String("pattern", pattern))
	}
}

// Modified CreateV2Product method
func (s *ProductService) CreateV2Product(ctx context.Context, params *model.CreateProductRequest, images []*multipart.FileHeader) error {
	timer := prometheus.NewTimer(productCreationDuration)
	defer timer.ObserveDuration()
	productCreationTotal.Inc()

	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		productCreationErrors.Inc()
		return s.logAndReturnError("failed to get user ID from context", err)
	}

	if err := s.verifyAdminStatus(ctx, userID); err != nil {
		productCreationErrors.Inc()
		return s.logAndReturnError("user is not authorized to create/update product", err)
	}

	// Define cache key using the new cache key generator
	cacheKey := ProductDetailKey(params.Slug)
	var existingProduct *model.ProductSchema

	// Use GetOrSet to handle cache retrieval and population, pass pointer to existingProduct
	err = s.cacheService.GetOrSet(ctx, cacheKey, &existingProduct, func() error {
		// Fetch from repository if cache miss
		fetchedProduct, err := s.productRepo.GetProductBySlug(ctx, params.Slug)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			s.logger.Error("failed to get product by slug", zap.Error(err))
			return fmt.Errorf("failed to get product by slug: %w", err)
		}
		existingProduct = fetchedProduct
		return nil
	})
	if err != nil {
		productCreationErrors.Inc()
		return s.logAndReturnError("failed to retrieve product from cache or repository", err)
	}

	var result error
	if existingProduct != nil {
		result = s.updateExistingProduct(ctx, existingProduct.ID, params, images)
	} else {
		result = s.createNewProduct(ctx, params, images)
	}

	if result != nil {
		productCreationErrors.Inc()
	} else {
		// Invalidate all product caches upon successful creation/update
		s.invalidateProductCache(ctx)
	}
	return result
}

func (s *ProductService) updateExistingProduct(ctx context.Context, productID uuid.UUID, params *model.CreateProductRequest, images []*multipart.FileHeader) error {
	err := s.productRepo.UpdateProduct(ctx, s.prepareUpdateProductParams(ctx, productID, params))
	if err != nil {
		return s.logAndReturnError("failed to update product", err)
	}
	return s.handleProductAssets(ctx, productID, params, images)
}

func (s *ProductService) createNewProduct(ctx context.Context, params *model.CreateProductRequest, images []*multipart.FileHeader) error {
	newProduct, err := s.productRepo.CreateProduct(ctx, s.prepareCreateProductParams(ctx, params))
	if err != nil {
		return s.logAndReturnError("failed to create product", err)
	}
	return s.handleProductAssets(ctx, newProduct.ID, params, images)
}

func (s *ProductService) prepareUpdateProductParams(ctx context.Context, productID uuid.UUID, params *model.CreateProductRequest) database.UpdateProductParams {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		s.logger.Error("failed to get user ID from context", zap.Error(err))
	}

	return database.UpdateProductParams{
		ID:          productID,
		Name:        strings.TrimSpace(params.Name),
		Description: sql.NullString{String: strings.TrimSpace(params.Description), Valid: true},
		RateToKes:   strconv.FormatFloat(params.Price, 'f', 2, 64),
		RateToKes_2: strconv.FormatFloat(params.PricePerUnit, 'f', 2, 64),
		Stock:       sql.NullInt32{Int32: int32(params.Stock), Valid: true},
		PartNumber:  params.PartNumber,
		MetaTitle: sql.NullString{
			String: params.MetaTitle,
			Valid:  params.MetaTitle != "",
		},
		MetaDescription: sql.NullString{
			String: params.MetaDescription,
			Valid:  params.MetaDescription != "",
		},
		Status:     params.Status,
		CategoryID: params.CategoryID,
		UpdatedBy: uuid.NullUUID{
			UUID:  userID,
			Valid: true,
		},
	}
}

func (s *ProductService) prepareCreateProductParams(ctx context.Context, params *model.CreateProductRequest) database.CreateProductParams {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		s.logger.Error("failed to get user ID from context", zap.Error(err))
	}

	return database.CreateProductParams{
		Name:        params.Name,
		Description: sql.NullString{String: params.Description, Valid: true},
		RateToKes:   strconv.FormatFloat(params.Price, 'f', 2, 64),
		RateToKes_2: strconv.FormatFloat(params.PricePerUnit, 'f', 2, 64),
		Stock:       sql.NullInt32{Int32: int32(params.Stock), Valid: true},
		Status:      params.Status,
		CategoryID:  params.CategoryID,
		MetaTitle: sql.NullString{
			String: params.MetaTitle,
			Valid:  params.MetaTitle != "",
		},
		MetaDescription: sql.NullString{
			String: params.MetaDescription,
			Valid:  params.MetaDescription != "",
		},
		PartNumber: params.PartNumber,
		CreatedBy: uuid.NullUUID{
			UUID:  userID,
			Valid: true,
		},
		UpdatedBy: uuid.NullUUID{
			UUID:  uuid.Nil,
			Valid: false,
		},
	}
}

func (s *ProductService) handleProductAssets(ctx context.Context, productID uuid.UUID, params *model.CreateProductRequest, images []*multipart.FileHeader) error {
	// First upload new images so they are inserted into the database.
	if err := s.uploadAndCreateProductImages(ctx, productID, images); err != nil {
		return err
	}

	// Then update the image positions according to the ordering provided in the request.
	// Ensure that params.ImageUrls is an ordered array of URLs (or file names) as desired.
	if err := s.updateProductImages(ctx, productID, params.ImageUrls); err != nil {
		return err
	}

	// Finally, upsert product specifications.
	return s.upsertProductSpecifications(ctx, productID, params.Specifications)
}

func (s *ProductService) updateProductImages(ctx context.Context, productID uuid.UUID, imageUrls []string) error {
	if len(imageUrls) == 0 {
		return nil
	}

	fileNames := s.extractFileNamesFromUrls(imageUrls)
	if len(fileNames) == 0 {
		return nil
	}

	return s.productImageRepo.UpdateProductImageUrls(ctx, productID, fileNames)
}

func (s *ProductService) uploadAndCreateProductImages(ctx context.Context, productID uuid.UUID, images []*multipart.FileHeader) error {
	if len(images) == 0 {
		return nil
	}

	uploadedFiles, err := utils.UploadMultipleFilesToS3(ctx, images, s.s3Client, s.config.AWSBucketName, "product-images")
	if err != nil {
		return s.logAndReturnError("failed to upload product images", err)
	}

	for _, image := range uploadedFiles {
		if image == "" {
			continue
		}
		_, err = s.productImageRepo.CreateProductImage(ctx, database.CreateProductImageParams{
			ProductID: uuid.NullUUID{UUID: productID, Valid: true},
			ImageUrl:  image,
		})
		if err != nil {
			return s.logAndReturnError("failed to create product image", err)
		}
	}
	return nil
}

func (s *ProductService) upsertProductSpecifications(ctx context.Context, productID uuid.UUID, specs []model.Specification) error {
	for _, spec := range specs {
		_, err := s.productSpecificationRepo.UpsertProductSpecification(ctx, database.UpsertProductSpecificationParams{
			ProductID: uuid.NullUUID{UUID: productID, Valid: true},
			SpecName:  spec.Name,
			SpecValue: spec.Value,
		})
		if err != nil {
			return s.logAndReturnError("failed to upsert product specification", err)
		}
	}
	return nil
}

func (s *ProductService) extractFileNamesFromUrls(urls []string) []string {
	var fileNames []string
	for _, u := range urls {
		fileName := s.getFilePathFromS3URL(u)
		if fileName != "" {
			fileNames = append(fileNames, fileName)
		}
	}
	return fileNames
}

func (s *ProductService) logAndReturnError(message string, err error) error {
	s.logger.Error(message, zap.Error(err))
	return fmt.Errorf("%s: %w", message, err)
}

// Modified DeleteProduct method
func (s *ProductService) DeleteProduct(ctx context.Context, slug string) error {
	// Get the product by slug
	product, err := s.productRepo.GetProductBySlug(ctx, slug)
	if err != nil {
		s.logger.Error("failed to get product by slug", zap.Error(err))
		return fmt.Errorf("failed to get product by slug: %w", err)
	}

	// Get the product images
	imageKeys, err := s.productImageRepo.GetImageKeysByProductID(ctx, product.ID)
	if err != nil {
		s.logger.Error("failed to get product images", zap.Error(err))
		return fmt.Errorf("failed to get product images: %w", err)
	}

	// Delete the product
	err = s.productRepo.DeleteProductByID(ctx, product.ID)
	if err != nil {
		s.logger.Error("failed to delete product", zap.Error(err))
		return fmt.Errorf("failed to delete product: %w", err)
	}

	// Delete the product images from S3
	if err := utils.DeleteMultipleFilesFromS3(ctx, s.s3Client, s.config.AWSBucketName, imageKeys); err != nil {
		s.logger.Error("failed to delete product images", zap.Error(err))
		return fmt.Errorf("failed to delete product images: %w", err)
	}

	// Invalidate all product caches upon deletion
	s.invalidateProductCache(ctx)
	return nil
}

// Modified ArchiveProduct method
func (s *ProductService) ArchiveProduct(ctx context.Context, slug string) error {
	// Get the product by slug
	product, err := s.productRepo.GetProductBySlug(ctx, slug)
	if err != nil {
		s.logger.Error("failed to get product by slug", zap.Error(err))
		return fmt.Errorf("failed to get product by slug: %w", err)
	}

	// Archive the product
	err = s.productRepo.ArchiveProductByID(ctx, product.ID)
	if err != nil {
		s.logger.Error("failed to archive product", zap.Error(err))
		return fmt.Errorf("failed to archive product: %w", err)
	}

	// Invalidate cache after archiving
	s.invalidateProductCache(ctx)
	return nil
}

// Modified ArchiveProducts method
func (s *ProductService) ArchiveProducts(ctx context.Context, slugs []string) error {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		err := app_errors.NewUnauthorizedUserError()
		s.logger.Error("failed to get user id from context", zap.Error(err))
		return err
	}

	isAdmin, err := s.userRepo.IsAdmin(ctx, userID)
	if err != nil {
		err := app_errors.NewUnauthorizedUserError()
		s.logger.Error("failed to check if user is admin", zap.Error(err))
		return err
	}

	if !isAdmin {
		err := app_errors.NewUnauthorizedUserError()
		s.logger.Error("user is not authorized to archive products", zap.Error(err))
		return err
	}

	err = s.productRepo.ArchiveProductsBySlugs(ctx, userID, slugs)
	if err != nil {
		s.logger.Error("failed to archive products", zap.Error(err))
		return fmt.Errorf("failed to archive products: %w", err)
	}

	// Invalidate all product caches after archiving multiple products
	s.invalidateProductCache(ctx)
	return nil
}

// Modified ActivateProducts method
func (s *ProductService) ActivateProducts(ctx context.Context, slugs []string) error {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		err := app_errors.NewUnauthorizedUserError()
		s.logger.Error("failed to get user id from context", zap.Error(err))
		return err
	}

	isAdmin, err := s.userRepo.IsAdmin(ctx, userID)
	if err != nil {
		err := app_errors.NewUnauthorizedUserError()
		s.logger.Error("failed to check if user is admin", zap.Error(err))
		return err
	}

	if !isAdmin {
		err := app_errors.NewUnauthorizedUserError()
		s.logger.Error("user is not authorized to activate products", zap.Error(err))
		return err
	}

	err = s.productRepo.ActivateProductsBySlugs(ctx, userID, slugs)
	if err != nil {
		s.logger.Error("failed to activate products", zap.Error(err))
		return fmt.Errorf("failed to activate products: %w", err)
	}

	// Invalidate cache after activating products
	s.invalidateProductCache(ctx)
	return nil
}

// Modified DeleteProducts method
func (s *ProductService) DeleteProducts(ctx context.Context, slugs []string) error {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		err := app_errors.NewUnauthorizedUserError()
		s.logger.Error("failed to get user id from context", zap.Error(err))
		return err
	}

	isAdmin, err := s.userRepo.IsAdmin(ctx, userID)
	if err != nil {
		err := app_errors.NewUnauthorizedUserError()
		s.logger.Error("failed to check if user is admin", zap.Error(err))
		return err
	}

	if !isAdmin {
		err := app_errors.NewUnauthorizedUserError()
		s.logger.Error("user is not authorized to delete products", zap.Error(err))
		return err
	}

	err = s.productRepo.DeleteProductsBySlugs(ctx, slugs)
	if err != nil {
		s.logger.Error("failed to delete products", zap.Error(err))
		return fmt.Errorf("failed to delete products: %w", err)
	}

	// Invalidate all product caches after deletion
	s.invalidateProductCache(ctx)
	return nil
}

// Modified DraftProducts method
func (s *ProductService) DraftProducts(ctx context.Context, slugs []string) error {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		err := app_errors.NewUnauthorizedUserError()
		s.logger.Error("failed to get user id from context", zap.Error(err))
		return err
	}

	isAdmin, err := s.userRepo.IsAdmin(ctx, userID)
	if err != nil {
		err := app_errors.NewUnauthorizedUserError()
		s.logger.Error("failed to check if user is admin", zap.Error(err))
		return err
	}

	if !isAdmin {
		err := app_errors.NewUnauthorizedUserError()
		s.logger.Error("user is not authorized to draft products", zap.Error(err))
		return err
	}

	err = s.productRepo.DraftProductsBySlugs(ctx, userID, slugs)
	if err != nil {
		s.logger.Error("failed to draft products", zap.Error(err))
		return fmt.Errorf("failed to draft products: %w", err)
	}

	// Invalidate cache after drafting products
	s.invalidateProductCache(ctx)
	return nil
}

// GetProductImagesBySlug retrieves all product images by slug with caching.
func (s *ProductService) GetProductImagesBySlug(ctx context.Context, slug string) ([]string, error) {
	cacheKey := ProductImagesKey(slug)

	var images []string

	// Define the fetch function to retrieve images from the repository
	fetchFunc := func() error {
		// Get the product by slug
		product, err := s.productRepo.GetProductBySlug(ctx, slug)
		if err != nil {
			s.logger.Error("Failed to get product by slug", zap.Error(err))
			return fmt.Errorf("failed to get product by slug: %w", err)
		}

		// Get the product images
		filePaths, err := s.productImageRepo.GetImageKeysByProductID(ctx, product.ID)
		if err != nil {
			s.logger.Error("Failed to get product images", zap.Error(err))
			return fmt.Errorf("failed to get product images: %w", err)
		}

		// Generate the image URLs
		for _, path := range filePaths {
			images = append(images, s.constructS3URL(path))
		}

		return nil
	}

	// Attempt to get the images from the cache or fetch and set them
	err := s.cacheService.GetOrSet(ctx, cacheKey, &images, fetchFunc)
	if err != nil {
		// If rate limit exceeded or other cache errors, proceed without cache
		s.logger.Warn("Cache GetOrSet failed, proceeding without cache", zap.Error(err))
		// Fallback: directly fetch from repository
		if err := fetchFunc(); err != nil {
			return nil, err
		}
	}

	s.logger.Debug("Product images retrieved from cache or database", zap.String("slug", slug))
	return images, nil
}

// GetProductPricingBySlug retrieves the product pricing by slug with caching.
func (s *ProductService) GetProductPricingBySlug(ctx context.Context, slug string) (*model.ProductPricing, error) {
	cacheKey := ProductPricingKey(slug)

	var pricing model.ProductPricing

	// Define the fetch function to retrieve pricing from the repository
	fetchFunc := func() error {
		// Get the product by slug
		product, err := s.productRepo.GetProductBySlug(ctx, slug)
		if err != nil {
			s.logger.Error("failed to get product by slug", zap.Error(err))
			return fmt.Errorf("failed to get product by slug: %w", err)
		}

		// Get the product pricing
		fetchedPricing, err := s.productRepo.GetProductPricingByProductID(ctx, product.ID)
		if err != nil {
			s.logger.Error("failed to get product pricing", zap.Error(err))
			return fmt.Errorf("failed to get product pricing: %w", err)
		}

		// Generate the image URL
		if fetchedPricing.ImageUrl != "" {
			fetchedPricing.ImageUrl = s.constructS3URL(fetchedPricing.ImageUrl)
		}

		// Assign to the destination
		pricing = *fetchedPricing
		return nil
	}

	// Attempt to get the pricing from the cache or fetch and set it
	err := s.cacheService.GetOrSet(ctx, cacheKey, &pricing, fetchFunc)
	if err != nil {
		// If rate limit exceeded or other cache errors, proceed without cache
		s.logger.Warn("cache.GetOrSet failed, proceeding without cache", zap.Error(err))
		// Fallback: directly fetch from repository
		if err := fetchFunc(); err != nil {
			return nil, err
		}
	}

	return &pricing, nil
}

// GetProductSpecsBySlug retrieves the product specifications by slug with caching.
func (s *ProductService) GetProductSpecsBySlug(ctx context.Context, slug string) (*model.ProductSpecs, error) {
	cacheKey := ProductSpecsKey(slug)

	var specs model.ProductSpecs

	// Define the fetch function to retrieve specs from the repository
	fetchFunc := func() error {
		// Get the product by slug
		product, err := s.productRepo.GetProductBySlug(ctx, slug)
		if err != nil {
			s.logger.Error("failed to get product by slug", zap.Error(err))
			return fmt.Errorf("failed to get product by slug: %w", err)
		}

		// Get the product specifications
		fetchedSpecs, err := s.productRepo.GetProductSpecsByID(ctx, product.ID)
		if err != nil {
			s.logger.Error("failed to get product specifications", zap.Error(err))
			return fmt.Errorf("failed to get product specifications: %w", err)
		}

		// Assign to the destination
		specs = *fetchedSpecs
		return nil
	}

	// Attempt to get the specs from the cache or fetch and set it
	err := s.cacheService.GetOrSet(ctx, cacheKey, &specs, fetchFunc)
	if err != nil {
		// If rate limit exceeded or other cache errors, proceed without cache
		s.logger.Warn("cache.GetOrSet failed, proceeding without cache", zap.Error(err))
		// Fallback: directly fetch from repository
		if err := fetchFunc(); err != nil {
			return nil, err
		}
	}

	return &specs, nil
}

// GetProductCart retrieves the product cart by slug
func (s *ProductService) GetProductCart(ctx context.Context, slug string) (*model.ProductCart, error) {
	cacheKey := ProductCartKey(slug)

	var cart model.ProductCart

	// Define the fetch function to retrieve cart from the repository
	fetchFunc := func() error {

		// Get the product cart
		fetchedProduct, err := s.productRepo.GetProductCartByProductSlug(ctx, slug)
		if err != nil {
			s.logger.Error("failed to get product cart", zap.Error(err))
			return fmt.Errorf("failed to get product cart: %w", err)
		}

		// Assign to the destination
		cart = *fetchedProduct
		return nil
	}

	// Attempt to get the cart from the cache or fetch and set it
	err := s.cacheService.GetOrSet(ctx, cacheKey, &cart, fetchFunc)
	if err != nil {
		// If rate limit exceeded or other cache errors, proceed without cache
		s.logger.Warn("cache.GetOrSet failed, proceeding without cache", zap.Error(err))
		// Fallback: directly fetch from repository
		if err := fetchFunc(); err != nil {
			return nil, err
		}
	}

	return &cart, nil
}
