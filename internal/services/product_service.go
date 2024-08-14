package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"weblineBackend/internal/app_errors"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
	"weblineBackend/pkg/utils"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProductService struct {
	productRepo              *repository.ProductRepository
	productVariantRepo       *repository.ProductVariantRepository
	productImageRepo         *repository.ProductImageRepository
	productSpecificationRepo *repository.ProductSpecificationRepository
	categoryRepo             *repository.CategoryRepository
	productColorRepo         *repository.ProductColourRepository
	productOptionRepo        *repository.ProductOptionRepository
	productSizeRepo          *repository.ProductSizeRepository
	discountRepo             *repository.DiscountRepository
	userRepo                 *repository.UserRepository
	exchangeRateRepo         repository.ExchangeRateRepository
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
	productColorRepo *repository.ProductColourRepository,
	productOptionRepo *repository.ProductOptionRepository,
	productSizeRepo *repository.ProductSizeRepository,
	discountRepo *repository.DiscountRepository,
	userRepo *repository.UserRepository,
	exchangeRateRepo repository.ExchangeRateRepository,
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
		productColorRepo:         productColorRepo,
		productSizeRepo:          productSizeRepo,
		discountRepo:             discountRepo,
		userRepo:                 userRepo,
		exchangeRateRepo:         exchangeRateRepo,
		logger:                   logger,
		config:                   config,
		s3Client:                 s3Client,
	}
}

// GetProductBySlug retrieves a product by its slug
func (s *ProductService) GetProductBySlug(ctx context.Context, slug string) (model.ProductDetail, error) {

	// Get the product from the repository
	product, err := s.productRepo.GetProductBySlug(ctx, slug)
	if err != nil {
		s.logger.Error("failed to get product by slug", zap.String("slug", slug), zap.Error(err))
		return model.ProductDetail{}, fmt.Errorf("failed to get product by slug: %w", err)
	}

	return model.ProductDetail{

		Name: product.Name,
		Slug: product.Slug,
	}, nil
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

	return priceToKES, nil
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

func (s *ProductService) getProductSpecifications(ctx context.Context, productID uuid.UUID) ([]model.ProductSpecification, error) {
	productSpecs, err := s.productSpecificationRepo.ListProductSpecificationsByProductID(ctx, uuid.NullUUID{UUID: productID, Valid: true})
	if err != nil {
		s.logger.Error("failed to get product specifications", zap.Error(err))
		return nil, fmt.Errorf("failed to get product specifications: %w", err)
	}

	var specs []model.ProductSpecification
	for _, spec := range productSpecs {
		specs = append(specs, model.ProductSpecification{
			ID:        spec.ID,
			SpecName:  spec.SpecName,
			SpecValue: spec.SpecValue,
		})
	}

	return specs, nil
}

func (s *ProductService) getProductImages(ctx context.Context, productID uuid.UUID) ([]model.ProductImage, error) {
	productImages, err := s.productImageRepo.ListProductImagesByProductID(ctx, uuid.NullUUID{UUID: productID, Valid: true})
	if err != nil {
		s.logger.Error("failed to get product images", zap.Error(err))
		return nil, fmt.Errorf("failed to get product images: %w", err)
	}

	var images []model.ProductImage
	for _, image := range productImages {
		images = append(images, model.ProductImage{
			ID:        image.ID,
			ProductID: image.ProductID.UUID.String(),
			S3URL:     s.constructS3URL(image.ImageUrl),
		})
	}

	return images, nil
}

func (s *ProductService) getProductColors(ctx context.Context, productID uuid.UUID) ([]model.ProductColor, error) {
	productColors, err := s.productColorRepo.ListProductColorsByProductID(ctx, uuid.NullUUID{UUID: productID, Valid: true})
	if err != nil {
		s.logger.Error("failed to get product colors", zap.Error(err))
		return nil, fmt.Errorf("failed to get product colors: %w", err)
	}

	var colors []model.ProductColor
	for _, color := range productColors {
		colors = append(colors, model.ProductColor{
			ID:        color.ID,
			ColorName: color.ColorName,
		})
	}

	return colors, nil
}

func (s *ProductService) getProductOptions(ctx context.Context, productID uuid.UUID) ([]model.ProductOption, error) {
	productOptions, err := s.productOptionRepo.ListProductOptionsByProductID(ctx, uuid.NullUUID{UUID: productID, Valid: true})
	if err != nil {
		s.logger.Error("failed to get product options", zap.Error(err))
		return nil, fmt.Errorf("failed to get product options: %w", err)
	}

	var options []model.ProductOption
	for _, option := range productOptions {
		values, err := s.productOptionRepo.ListProductOptionValuesByOptionID(ctx, uuid.NullUUID{UUID: option.ID, Valid: true})
		if err != nil {
			s.logger.Error("failed to get product option values", zap.Error(err))
			return nil, fmt.Errorf("failed to get product option values: %w", err)
		}

		var optionValues []model.ProductOptionValue
		for _, value := range values {
			additionalPrice, err := strconv.ParseFloat(value.AdditionalPrice.String, 64)
			if err != nil {
				s.logger.Error("failed to parse additional price", zap.Error(err))
				return nil, fmt.Errorf("failed to parse additional price: %w", err)
			}

			optionValues = append(optionValues, model.ProductOptionValue{
				ID:              value.ID,
				ValueName:       value.ValueName,
				AdditionalPrice: additionalPrice,
			})
		}

		options = append(options, model.ProductOption{
			ID:           option.ID,
			OptionName:   option.OptionName,
			OptionValues: optionValues,
		})
	}

	return options, nil
}

// GetProductsByCategoryID retrieves products by their category ID
func (s *ProductService) GetProductsByCategoryID(ctx context.Context, categoryID string, pageSize int32, page int32) (model.PaginationResult[[]model.Product], error) {
	// Parse the category ID
	categoryIDValue, err := uuid.Parse(categoryID)
	if err != nil {
		s.logger.Error("invalid category ID format", zap.String("categoryID", categoryID), zap.Error(err))
		return model.PaginationResult[[]model.Product]{}, fmt.Errorf("invalid category ID format: %w", err)
	}

	// Get total number of products by category ID
	totalProductsByCategory, err := s.productRepo.CountProductsByParentCategoryID(ctx, categoryIDValue)
	if err != nil {
		s.logger.Error("failed to count products by category ID", zap.Error(err))
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
				s.logger.Error("failed to get products by category ID", zap.Error(err))
				return nil, fmt.Errorf("failed to get products by category ID: %w", err)
			}

			log.Println(products)

			return s.mapProductsToModel(ctx, products)
		},
	)

	if err != nil {
		s.logger.Error("failed to paginate products by category ID", zap.Error(err))
		return model.PaginationResult[[]model.Product]{}, fmt.Errorf("failed to paginate products by category ID: %w", err)
	}

	return *paginatedProductsByCategory, nil
}

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

		discountPercentage, err := s.getProductDiscountPercentage(ctx, product.ID)
		if err != nil {
			return nil, err
		}

		imageUrl := ""
		if len(productImages) > 0 {
			imageUrl = s.constructS3URL(productImages[0].ImageUrl)
		}

		productSchemas = append(productSchemas, model.Product{
			ID:              product.ID,
			Name:            product.Name,
			Description:     product.Description,
			Price:           fmt.Sprintf("%.2f", price),
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

		productsQueryResult = append(productsQueryResult, model.ProductQueryResult{
			ID:              product.ID,
			Name:            product.Name,
			Price:           fmt.Sprintf("%.2f", price),
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

//// GetProductsByParentCategoryID retrieves products by their parent category ID
//func (s *ProductService) GetProductsByParentCategoryID(ctx context.Context, parentCategoryID string) ([]model.ProductDetail, error) {
//	// Parse the parent category ID
//	parentCategoryUUID, err := uuid.Parse(parentCategoryID)
//	if err != nil {
//		s.logger.Error("invalid parent category ID format", zap.String("parentCategoryID", parentCategoryID), zap.Error(err))
//		return nil, fmt.Errorf("invalid parent category ID format: %w", err)
//	}
//
//	// Get products by category ID
//	productsByCategory, err := s.productRepo.GetProductsByParentCategoryID(ctx, parentCategoryUUID)
//	if err != nil {
//		s.logger.Error("failed to get products by category ID", zap.Error(err))
//		return nil, fmt.Errorf("failed to get products by category ID: %w", err)
//	}
//
//	return s.mapProductsToDetail(ctx, productsByCategory)
//}

//func (s *ProductService) mapProductsToDetail(ctx context.Context, products []model.ProductSchema) ([]model.ProductDetail, error) {
//	var productDetails []model.ProductDetail
//
//	for _, product := range products {
//		// Get additional product information
//		specs, err := s.getProductSpecifications(ctx, product.ID)
//		if err != nil {
//			return nil, err
//		}
//
//		images, err := s.getProductImages(ctx, product.ID)
//		if err != nil {
//			return nil, err
//		}
//
//		colors, err := s.getProductColors(ctx, product.ID)
//		if err != nil {
//			return nil, err
//		}
//
//		options, err := s.getProductOptions(ctx, product.ID)
//		if err != nil {
//			return nil, err
//		}
//
//		discountPercentage, err := s.getProductDiscountPercentage(ctx, product.ID)
//		if err != nil {
//			return nil, err
//		}
//
//		// Calculate price to KES
//		price, err := s.calculatePriceToKES(ctx, product.USD)
//		if err != nil {
//			s.logger.Error("failed to calculate price to KES", zap.Error(err))
//			return nil, err
//		}
//
//		productDetails = append(productDetails, model.ProductDetail{
//			ID:              product.ID,
//			Name:            product.Name,
//			Description:     product.Description,
//			Price:           fmt.Sprintf("%.2f", price),
//			Stock:           product.Stock,
//			CategoryName:    product.CategoryName,
//			Featured:        product.Featured,
//			Colors:          colors,
//			Specifications:  specs,
//			Options:         options,
//			Images:          images,
//			DiscountPercent: discountPercentage,
//		})
//	}
//
//	return productDetails, nil
//}

// CreateProductColor creates a new product color
func (s *ProductService) CreateProductColor(
	ctx context.Context,
	productID,
	color string,
) (*database.CreateProductColorRow, error) {
	// Parse the product ID
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("invalid product ID", zap.Error(err))
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	// Prepare parameters for creating product color
	params := database.CreateProductColorParams{
		ProductID: uuid.NullUUID{UUID: productUUID, Valid: true},
		ColorName: color,
	}

	// Create product color
	productColor, err := s.productColorRepo.CreateProductColor(ctx, params)
	if err != nil {
		s.logger.Error("failed to create product color", zap.Error(err))
		return nil, fmt.Errorf("failed to create product color: %w", err)
	}

	return productColor, nil
}

// ListProductColorsByProductID lists product colors by product ID
func (s *ProductService) ListProductColorsByProductID(
	ctx context.Context,
	productID string,
) ([]database.ListProductColorsByProductIDRow, error) {
	// Parse the product ID
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("invalid product ID", zap.Error(err))
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	// List product colors by product ID
	productColors, err := s.productColorRepo.ListProductColorsByProductID(ctx, uuid.NullUUID{UUID: productUUID, Valid: true})
	if err != nil {
		s.logger.Error("failed to list product colors", zap.Error(err))
		return nil, fmt.Errorf("failed to list product colors: %w", err)
	}

	return productColors, nil
}

// UpdateProductColor updates an existing product color
func (s *ProductService) UpdateProductColor(
	ctx context.Context,
	id string,
	color string,
) (*database.UpdateProductColorRow, error) {
	// Parse the product color ID
	colorUUID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid product color ID", zap.Error(err))
		return nil, fmt.Errorf("invalid product color ID: %w", err)
	}

	// Prepare parameters for updating product color
	params := database.UpdateProductColorParams{
		ID:        colorUUID,
		ColorName: color,
	}

	// Update product color
	productColor, err := s.productColorRepo.UpdateProductColor(ctx, params)
	if err != nil {
		s.logger.Error("failed to update product color", zap.Error(err))
		return nil, fmt.Errorf("failed to update product color: %w", err)
	}

	return productColor, nil
}

// DeleteProductColor deletes a product color from the database
func (s *ProductService) DeleteProductColor(ctx context.Context, id string) error {
	// Parse the product color ID
	colorUUID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid product color ID", zap.Error(err))
		return fmt.Errorf("invalid product color ID: %w", err)
	}

	// Delete product color
	err = s.productColorRepo.DeleteProductColor(ctx, colorUUID)
	if err != nil {
		s.logger.Error("failed to delete product color", zap.Error(err))
		return fmt.Errorf("failed to delete product color: %w", err)
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

// GetProductsByFilters retrieves products by filters and sort order.
func (s *ProductService) GetProductsByFilters(
	ctx context.Context,
	categoryID uuid.UUID,
	categoryNames, colorNames, processorNames, storageNames, sizes []string,
	priceFrom, priceTo float64,
	sortOrder string,
) ([]model.FilterProduct, error) {

	params := model.UnifiedParams{
		ID:             categoryID,
		CategoryNames:  categoryNames,
		ColorNames:     colorNames,
		ProcessorNames: processorNames,
		StorageNames:   storageNames,
		Sizes:          sizes,
		PriceFrom:      priceFrom,
		PriceTo:        priceTo,
		SortOrder:      sortOrder,
	}

	filterProducts, err := s.productRepo.GetProductsByFilters(ctx, params)
	if err != nil {
		s.logger.Error("failed to get products by filters", zap.Error(err))
		return nil, fmt.Errorf("failed to get products by filters: %w", err)
	}

	for i := range filterProducts {
		productImages, err := s.getProductImages(ctx, filterProducts[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product images: %w", err)
		}

		discountPercent, err := s.getProductDiscountPercentage(ctx, filterProducts[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product discount percentage: %w", err)
		}

		if len(productImages) > 0 {
			filterProducts[i].ImageURL = productImages[0].S3URL
		}
		filterProducts[i].DiscountPercent = discountPercent
	}

	return filterProducts, nil
}

type FilterOptionsByType map[string][]string

// GetFilterOptionsByCategoryName retrieves filter options for products by category name
func (s *ProductService) GetFilterOptionsByCategoryName(ctx context.Context, categoryName string) (FilterOptionsByType, error) {
	filterOptions, err := s.categoryRepo.GetFilterOptionsByCategoryName(ctx, categoryName)
	if err != nil {
		s.logger.Error("failed to get filter options by category name", zap.Error(err))
		return nil, fmt.Errorf("failed to get filter options by category name: %w", err)
	}

	groupedOptions := make(FilterOptionsByType)
	for _, option := range filterOptions {
		if _, exists := groupedOptions[option.FilterType]; !exists {
			groupedOptions[option.FilterType] = []string{}
		}
		groupedOptions[option.FilterType] = append(groupedOptions[option.FilterType], option.FilterOption)
	}

	return groupedOptions, nil
}

// GetFilterOptionsByCategoryID retrieves filter options for products by category ID
func (s *ProductService) GetFilterOptionsByCategoryID(ctx context.Context, categoryID uuid.UUID) (*model.ProductMetafields, error) {
	filterOptions, err := s.categoryRepo.GetFilterOptionsByCategoryID(ctx, categoryID)
	if err != nil {
		s.logger.Error("failed to get filter options by category name", zap.Error(err))
		return nil, fmt.Errorf("failed to get filter options by category name: %w", err)
	}

	return filterOptions, nil
}

type FilterOptions struct {
	Categories    []model.ProductCategoryFilterOption `json:"categories"`
	TotalProducts int64                               `json:"totalProducts"`
	Processor     []string                            `json:"processor"`
	Storage       []string                            `json:"storage"`
	Color         []string                            `json:"color"`
	Size          []string                            `json:"size"`
}

// GetFilterOptions retrieves filter options for products
func (s *ProductService) GetFilterOptions(ctx context.Context) (*FilterOptions, error) {

	filterOptions, err := s.productRepo.GetFilterOptions(ctx)
	if err != nil {
		s.logger.Error("failed to get filter options", zap.Error(err))
		return nil, fmt.Errorf("failed to get filter options: %w", err)
	}

	// Get available categories and their subcategories
	categories, err := s.categoryRepo.GetParentCategories(ctx)
	if err != nil {
		s.logger.Error("failed to get parent categories", zap.Error(err))
		return nil, fmt.Errorf("failed to get parent categories: %w", err)
	}

	// Get total number of products
	count, err := s.productRepo.CountProducts(ctx)
	if err != nil {
		s.logger.Error("failed to get total products", zap.Error(err))
		return nil, fmt.Errorf("failed to get total products: %w", err)
	}

	filters := &FilterOptions{
		TotalProducts: count,
	}

	// Group the filter options by their type
	var categoryOptions []model.ProductCategoryFilterOption
	for _, category := range categories {
		// Get the subcategories of each parent category
		subCategories, err := s.categoryRepo.GetCategoriesByParentID(ctx, uuid.NullUUID{UUID: category.ID, Valid: true})
		if err != nil {
			s.logger.Error("failed to get subcategories", zap.Error(err))
			return nil, fmt.Errorf("failed to get subcategories: %w", err)
		}

		var subCategoryOptions []model.ProductFilterOption
		for _, subCategory := range subCategories {
			subCategoryOptions = append(subCategoryOptions, model.ProductFilterOption{
				ID:   subCategory.ID,
				Name: subCategory.Name,
			})
		}

		categoryOptions = append(categoryOptions, model.ProductCategoryFilterOption{
			Title:         category.Name,
			Subcategories: subCategoryOptions,
		})
	}

	filters.Categories = categoryOptions

	for _, option := range filterOptions {
		switch option.FilterType {
		case "processor":
			filters.Processor = append(filters.Processor, option.FilterOption)
		case "storage":
			filters.Storage = append(filters.Storage, option.FilterOption)
		case "color":
			filters.Color = append(filters.Color, option.FilterOption)
		case "size":
			filters.Size = append(filters.Size, option.FilterOption)
		}
	}

	return filters, nil
}

// GetAllProductsByFilters retrieves all products by filters with specified sorting order
func (s *ProductService) GetAllProductsByFilters(
	ctx context.Context,
	categoryNames, colorNames, processorNames, storageNames, sizes []string,
	priceFrom, priceTo float64,
	page, pageSize int32,
	sortOrder string,
) (*model.PaginationResult[[]*model.Product], error) {
	// get the total number of products
	count, err := s.productRepo.GetTotalProductsByFilters(ctx, database.GetTotalProductsByFiltersParams{
		Column1:    categoryNames,
		Column4:    sizes,
		Column5:    colorNames,
		Column6:    processorNames,
		Column7:    storageNames,
		UsdPrice:   strconv.FormatFloat(priceFrom, 'f', -1, 64),
		UsdPrice_2: strconv.FormatFloat(priceTo, 'f', -1, 64),
	})
	if err != nil {
		s.logger.Error("failed to get total products by filters", zap.Error(err))
		return nil, fmt.Errorf("failed to get total products by filters: %w", err)
	}

	// get the paginated products by filters with specified sorting order
	paginatedProducts, err := utils.Paginate(
		s.config,
		count,
		page,
		pageSize,
		func(offset int32, limit int32) ([]*model.Product, error) {
			var filteredProducts []*model.Product
			switch sortOrder {
			case "price_asc":
				filteredProducts, err = s.productRepo.GetAllProductsByFiltersPriceAsc(ctx, database.GetAllProductsByFiltersPriceAscParams{
					Column1:    categoryNames,
					Column4:    sizes,
					Column5:    colorNames,
					Column8:    processorNames,
					Column9:    storageNames,
					UsdPrice:   strconv.FormatFloat(priceFrom, 'f', -1, 64),
					UsdPrice_2: strconv.FormatFloat(priceTo, 'f', -1, 64),
					Limit:      limit,
					Offset:     offset,
				})
			case "price_desc":
				filteredProducts, err = s.productRepo.GetAllProductsByFiltersPriceDesc(ctx, database.GetAllProductsByFiltersPriceDescParams{
					Column1:    categoryNames,
					Column4:    sizes,
					Column5:    colorNames,
					Column8:    processorNames,
					Column9:    storageNames,
					UsdPrice:   strconv.FormatFloat(priceFrom, 'f', -1, 64),
					UsdPrice_2: strconv.FormatFloat(priceTo, 'f', -1, 64),
					Limit:      limit,
					Offset:     offset,
				})
			case "name_asc":
				filteredProducts, err = s.productRepo.GetAllProductsByFiltersNameAsc(ctx, database.GetAllProductsByFiltersNameAscParams{
					Column1:    categoryNames,
					Column4:    sizes,
					Column5:    colorNames,
					Column8:    processorNames,
					Column9:    storageNames,
					UsdPrice:   strconv.FormatFloat(priceFrom, 'f', -1, 64),
					UsdPrice_2: strconv.FormatFloat(priceTo, 'f', -1, 64),
					Limit:      limit,
					Offset:     offset,
				})
			case "name_desc":
				filteredProducts, err = s.productRepo.GetAllProductsByFiltersNameDesc(ctx, database.GetAllProductsByFiltersNameDescParams{
					Column1:    categoryNames,
					Column4:    sizes,
					Column5:    colorNames,
					Column8:    processorNames,
					Column9:    storageNames,
					UsdPrice:   strconv.FormatFloat(priceFrom, 'f', -1, 64),
					UsdPrice_2: strconv.FormatFloat(priceTo, 'f', -1, 64),
					Limit:      limit,
					Offset:     offset,
				})
			case "newest":
				filteredProducts, err = s.productRepo.GetAllProductsByFiltersNewest(ctx, database.GetAllProductsByFiltersNewestParams{
					Column1:    categoryNames,
					Column4:    sizes,
					Column5:    colorNames,
					Column8:    processorNames,
					Column9:    storageNames,
					UsdPrice:   strconv.FormatFloat(priceFrom, 'f', -1, 64),
					UsdPrice_2: strconv.FormatFloat(priceTo, 'f', -1, 64),
					Limit:      limit,
					Offset:     offset,
				})
			case "oldest":
				filteredProducts, err = s.productRepo.GetAllProductsByFiltersOldest(ctx, database.GetAllProductsByFiltersOldestParams{
					Column1:    categoryNames,
					Column4:    sizes,
					Column5:    colorNames,
					Column8:    processorNames,
					Column9:    storageNames,
					UsdPrice:   strconv.FormatFloat(priceFrom, 'f', -1, 64),
					UsdPrice_2: strconv.FormatFloat(priceTo, 'f', -1, 64),
					Limit:      limit,
					Offset:     offset,
				})
			default:
				filteredProducts, err = s.productRepo.GetAllProductsByFiltersNewest(ctx, database.GetAllProductsByFiltersNewestParams{
					Column1:    categoryNames,
					Column4:    sizes,
					Column5:    colorNames,
					Column8:    processorNames,
					Column9:    storageNames,
					UsdPrice:   strconv.FormatFloat(priceFrom, 'f', -1, 64),
					UsdPrice_2: strconv.FormatFloat(priceTo, 'f', -1, 64),
					Limit:      limit,
					Offset:     offset,
				})
			}

			if err != nil {
				s.logger.Error("failed to get products by filters", zap.Error(err))
				return nil, fmt.Errorf("failed to get products by filters: %w", err)
			}

			for _, product := range filteredProducts {
				productImages, err := s.productImageRepo.ListProductImagesByProductID(ctx, uuid.NullUUID{UUID: product.ID, Valid: true})
				if err != nil {
					s.logger.Error("failed to get product images", zap.Error(err))
					return nil, fmt.Errorf("failed to get product images: %w", err)
				}

				discountPercentage, err := s.getProductDiscountPercentage(ctx, product.ID)
				if err != nil {
					return nil, err
				}

				if len(productImages) > 0 {
					product.ImageURL = s.constructS3URL(productImages[0].ImageUrl)
				}

				product.DiscountPercent = discountPercentage
			}

			return filteredProducts, nil
		},
	)
	if err != nil {
		s.logger.Error("failed to paginate products", zap.Error(err))
		return nil, fmt.Errorf("failed to paginate products: %w", err)
	}

	return paginatedProducts, nil
}

// GetAllProductSitemap retrieves all products for sitemap
func (s *ProductService) GetAllProductSitemap(ctx context.Context) ([]*model.ProductSitemap, error) {
	products, err := s.productRepo.ListProducts(ctx, database.ListProductsParams{
		Offset: 0,
		Limit:  100,
	})
	if err != nil {
		s.logger.Error("failed to get all product sitemap", zap.Error(err))
		return nil, fmt.Errorf("failed to get all product sitemap: %w", err)
	}

	productSitemap := make([]*model.ProductSitemap, 0, len(products))
	for _, product := range products {
		productSitemap = append(productSitemap, &model.ProductSitemap{
			ID:        product.ID,
			UpdatedAt: product.UpdatedAt.Time,
		})
	}

	return productSitemap, nil
}

// GetProducts retrieves all products
func (s *ProductService) GetProducts(ctx context.Context) ([]*model.V2Product, error) {
	products, err := s.productRepo.GetV2Products(ctx)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			s.logger.Error("no products found")
			return nil, err
		default:
			s.logger.Error("failed to get products", zap.Error(err))
			return nil, fmt.Errorf("failed to get products: %w", err)
		}
	}

	for _, product := range products {
		// update the product image URL
		if product.ImageURL != "" {
			product.ImageURL = s.constructS3URL(product.ImageURL)
		}
	}

	return products, nil
}

// GetProductDetail retrieves a product by slug
func (s *ProductService) GetProductDetail(ctx context.Context, slug string) (*model.V2ProductDetail, error) {
	product, err := s.productRepo.GetV2ProductDetailBySlug(ctx, slug)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			s.logger.Error("product not found", zap.String("slug", slug))
			return nil, fmt.Errorf("product not found: %w", err)
		default:
			s.logger.Error("failed to get product", zap.Error(err))
			return nil, fmt.Errorf("failed to get product: %w", err)
		}
	}

	var images []model.V2ProductImage
	if err := json.Unmarshal(product.Images, &images); err != nil {
		s.logger.Error("failed to unmarshal images", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal images: %w", err)
	}

	// update the product image URL
	for i := range images {
		if images[i].Url != "" {
			images[i].Url = s.constructS3URL(images[i].Url)
		}
	}

	updatedImages, err := json.Marshal(images)
	if err != nil {
		s.logger.Error("failed to marshal images", zap.Error(err))
		return nil, fmt.Errorf("failed to marshal images: %w", err)
	}

	product.Images = updatedImages
	return product, nil
}

func (s *ProductService) CreateV2Product(ctx context.Context, params *model.CreateProductRequest, images []*multipart.FileHeader) error {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return s.logAndReturnError("failed to get user ID from context", err)
	}

	if err := s.verifyAdminStatus(ctx, userID); err != nil {
		return s.logAndReturnError("user is not authorized to create/update product", err)
	}

	existingProduct, err := s.productRepo.GetProductBySlug(ctx, params.Slug)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return s.logAndReturnError("failed to get product by slug", err)
	}

	if existingProduct != nil {
		return s.updateExistingProduct(ctx, existingProduct.ID, params, images)
	}
	return s.createNewProduct(ctx, params, images)
}

func (s *ProductService) getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value("userId").(uuid.UUID)
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

func (s *ProductService) updateExistingProduct(ctx context.Context, productID uuid.UUID, params *model.CreateProductRequest, images []*multipart.FileHeader) error {
	err := s.productRepo.UpdateProduct(ctx, s.prepareUpdateProductParams(productID, params))
	if err != nil {
		return s.logAndReturnError("failed to update product", err)
	}
	return s.handleProductAssets(ctx, productID, params, images)
}

func (s *ProductService) createNewProduct(ctx context.Context, params *model.CreateProductRequest, images []*multipart.FileHeader) error {
	newProduct, err := s.productRepo.CreateProduct(ctx, s.prepareCreateProductParams(params))
	if err != nil {
		return s.logAndReturnError("failed to create product", err)
	}
	return s.handleProductAssets(ctx, newProduct.ID, params, images)
}

func (s *ProductService) prepareUpdateProductParams(productID uuid.UUID, params *model.CreateProductRequest) database.UpdateProductParams {
	return database.UpdateProductParams{
		ID:          productID,
		Name:        strings.TrimSpace(params.Name),
		Description: sql.NullString{String: strings.TrimSpace(params.Description), Valid: true},
		RateToKes:   fmt.Sprintf("%.0f", params.Price),
		Stock:       sql.NullInt32{Int32: int32(params.Stock), Valid: true},
		Status:      params.Status,
		CategoryID:  params.CategoryID,
		UpdatedBy: uuid.NullUUID{
			UUID:  uuid.Nil,
			Valid: false,
		},
	}
}

func (s *ProductService) prepareCreateProductParams(params *model.CreateProductRequest) database.CreateProductParams {
	return database.CreateProductParams{
		Name:        params.Name,
		Description: sql.NullString{String: params.Description, Valid: true},
		RateToKes:   fmt.Sprintf("%.0f", params.Price),
		Stock:       sql.NullInt32{Int32: int32(params.Stock), Valid: true},
		Status:      params.Status,
		CategoryID:  params.CategoryID,
		PartNumber:  params.PartNumber,
		CreatedBy: uuid.NullUUID{
			UUID:  uuid.Nil,
			Valid: false,
		},
		UpdatedBy: uuid.NullUUID{
			UUID:  uuid.Nil,
			Valid: false,
		},
	}
}

func (s *ProductService) handleProductAssets(ctx context.Context, productID uuid.UUID, params *model.CreateProductRequest, images []*multipart.FileHeader) error {
	if err := s.updateProductImages(ctx, productID, params.ImageUrls); err != nil {
		return err
	}
	if err := s.uploadAndCreateProductImages(ctx, productID, images); err != nil {
		return err
	}
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

// DeleteProduct deletes a product
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

	return nil
}

// ArchiveProduct archives a product
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

	return nil
}

// ArchiveProducts archives multiple products
func (s *ProductService) ArchiveProducts(ctx context.Context, slugs []string) error {
	// Get the user ID from the context
	userID, ok := ctx.Value("userId").(uuid.UUID)
	if !ok {
		err := app_errors.NewUnauthorizedUserError()
		s.logger.Error("failed to get user id from context", zap.Error(err))
		return err
	}

	// Check if user is admin
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

	// Archive the products
	err = s.productRepo.ArchiveProductsBySlugs(ctx, userID, slugs)
	if err != nil {
		s.logger.Error("failed to archive products", zap.Error(err))
		return fmt.Errorf("failed to archive products: %w", err)
	}

	return nil
}

// ActivateProducts activates multiple products
func (s *ProductService) ActivateProducts(ctx context.Context, slugs []string) error {
	// Get the user ID from the context
	userID, ok := ctx.Value("userId").(uuid.UUID)
	if !ok {
		err := app_errors.NewUnauthorizedUserError()
		s.logger.Error("failed to get user id from context", zap.Error(err))
		return err
	}

	// Check if user is admin
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

	// Activate the products
	err = s.productRepo.ActivateProductsBySlugs(ctx, userID, slugs)
	if err != nil {
		s.logger.Error("failed to activate products", zap.Error(err))
		return fmt.Errorf("failed to activate products: %w", err)
	}

	return nil
}

// DeleteProducts deletes multiple products
func (s *ProductService) DeleteProducts(ctx context.Context, slugs []string) error {
	// Get the user ID from the context
	userID, ok := ctx.Value("userId").(uuid.UUID)
	if !ok {
		err := app_errors.NewUnauthorizedUserError()
		s.logger.Error("failed to get user id from context", zap.Error(err))
		return err
	}

	// Check if user is admin
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

	// Delete the products
	err = s.productRepo.DeleteProductsBySlugs(ctx, slugs)
	if err != nil {
		s.logger.Error("failed to delete products", zap.Error(err))
		return fmt.Errorf("failed to delete products: %w", err)
	}

	return nil
}

// DraftProducts drafts multiple products
func (s *ProductService) DraftProducts(ctx context.Context, slugs []string) error {
	// Get the user ID from the context
	userID, ok := ctx.Value("userId").(uuid.UUID)
	if !ok {
		err := app_errors.NewUnauthorizedUserError()
		s.logger.Error("failed to get user id from context", zap.Error(err))
		return err
	}

	// Check if user is admin
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

	// Draft the products
	err = s.productRepo.DraftProductsBySlugs(ctx, userID, slugs)
	if err != nil {
		s.logger.Error("failed to draft products", zap.Error(err))
		return fmt.Errorf("failed to draft products: %w", err)
	}

	return nil
}

// GetProductImagesBySlug retrieves all product images by slug
func (s *ProductService) GetProductImagesBySlug(ctx context.Context, slug string) ([]string, error) {
	// Get the product by slug
	product, err := s.productRepo.GetProductBySlug(ctx, slug)
	if err != nil {
		s.logger.Error("failed to get product by slug", zap.Error(err))
		return nil, fmt.Errorf("failed to get product by slug: %w", err)
	}

	// Get the product images
	filePath, err := s.productImageRepo.GetImageKeysByProductID(ctx, product.ID)
	if err != nil {
		s.logger.Error("failed to get product images", zap.Error(err))
		return nil, fmt.Errorf("failed to get product images: %w", err)
	}

	var images []string
	for _, path := range filePath {
		images = append(images, s.constructS3URL(path))
	}

	return images, nil
}

// GetProductPricingBySlug retrieves the product pricing by slug
func (s *ProductService) GetProductPricingBySlug(ctx context.Context, slug string) (*model.ProductPricing, error) {
	// Get the product by slug
	product, err := s.productRepo.GetProductBySlug(ctx, slug)
	if err != nil {
		s.logger.Error("failed to get product by slug", zap.Error(err))
		return nil, fmt.Errorf("failed to get product by slug: %w", err)
	}

	// Get the product pricing
	pricing, err := s.productRepo.GetProductPricingByProductID(ctx, product.ID)
	if err != nil {
		s.logger.Error("failed to get product pricing", zap.Error(err))
		return nil, fmt.Errorf("failed to get product pricing: %w", err)
	}

	// generate the image url
	pricing.ImageUrl = s.constructS3URL(pricing.ImageUrl)

	return pricing, nil
}

// GetProductSpecsBySlug retrieves the product specifications by slug
func (s *ProductService) GetProductSpecsBySlug(ctx context.Context, slug string) (*model.ProductSpecs, error) {
	// Get the product by slug
	product, err := s.productRepo.GetProductBySlug(ctx, slug)
	if err != nil {
		s.logger.Error("failed to get product by slug", zap.Error(err))
		return nil, fmt.Errorf("failed to get product by slug: %w", err)
	}

	// Get the product specifications
	specs, err := s.productRepo.GetProductSpecsByID(ctx, product.ID)
	if err != nil {
		s.logger.Error("failed to get product specifications", zap.Error(err))
		return nil, fmt.Errorf("failed to get product specifications: %w", err)
	}

	return specs, nil
}
