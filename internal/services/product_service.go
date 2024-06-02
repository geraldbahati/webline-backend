package services

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
	"weblineBackend/pkg/utils"
)

type ProductService struct {
	productRepo              *repository.ProductRepository
	productVariantRepo       *repository.ProductVariantRepository
	productImageRepo         *repository.ProductImageRepository
	productSpecificationRepo *repository.ProductSpecificationRepository
	categoryRepo             *repository.CategoryRepository
	productColorRepo         *repository.ProductColourRepository
	productOptionRepo        *repository.ProductOptionRepository
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
		logger:                   logger,
		config:                   config,
		s3Client:                 s3Client,
	}
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(
	ctx context.Context,
	name string,
	description string,
	price float64,
	categoryID string,
	stock int32,
) (database.Product, error) {
	// get user from context
	userId, exists := ctx.Value("userId").(uuid.UUID)
	if !exists {
		s.logger.Error("failed to get user id from context")
		userId = uuid.Nil
	}

	// parsing
	descriptionValue := sql.NullString{String: description, Valid: description != ""}
	categoryIDValue := uuid.NullUUID{UUID: uuid.Nil, Valid: false}
	if categoryID != "" {
		if parsedUUID, err := uuid.Parse(categoryID); err == nil {
			categoryIDValue.UUID = parsedUUID
			categoryIDValue.Valid = true
		}
	}
	stockValue := sql.NullInt32{Int32: stock, Valid: stock > 0}
	creatorIDValue := uuid.NullUUID{UUID: userId, Valid: userId != uuid.Nil}

	params := database.CreateProductParams{
		Name:        name,
		Description: descriptionValue,
		Price:       strconv.FormatFloat(price, 'f', -1, 64),
		CategoryID:  categoryIDValue,
		Stock:       stockValue,
		CreatedBy:   creatorIDValue,
		UpdatedBy:   creatorIDValue,
	}

	// create product
	product, err := s.productRepo.CreateProduct(ctx, params)
	if err != nil {
		s.logger.Error("failed to create product", zap.Error(err))
		return database.Product{}, err
	}

	return product, nil
}

// GetProductByID retrieves a product by its ID
func (s *ProductService) GetProductByID(ctx context.Context, productID string) (database.Product, error) {
	// Parse the product ID
	id, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("invalid product ID format", zap.String("productID", productID), zap.Error(err))
		return database.Product{}, fmt.Errorf("invalid product ID format: %w", err)
	}

	// Get the product from the repository
	product, err := s.productRepo.GetProductByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get product by ID", zap.String("productID", productID), zap.Error(err))
		return database.Product{}, fmt.Errorf("failed to get product by ID: %w", err)
	}

	return product, nil
}

// ListProducts retrieves all products from the database
func (s *ProductService) ListProducts(ctx context.Context, page int32, pageSize int32) (model.PaginationResult[[]database.Product], error) {
	// Get total number of products
	totalProducts, err := s.productRepo.CountProducts(ctx)
	if err != nil {
		s.logger.Error("failed to count products", zap.Error(err))
		return model.PaginationResult[[]database.Product]{}, fmt.Errorf("failed to count products: %w", err)
	}

	// Get all products
	paginatedProducts, err := utils.Paginate(
		s.config,
		totalProducts,
		page,
		pageSize,
		func(offset int32, limit int32) ([]database.Product, error) {
			products, err := s.productRepo.ListProducts(ctx, limit, offset)
			if err != nil {
				s.logger.Error("failed to list products", zap.Error(err))
				return nil, fmt.Errorf("failed to list products: %w", err)
			}
			return products, nil
		},
	)

	return *paginatedProducts, nil
}

// UpdateProduct updates an existing product
func (s *ProductService) UpdateProduct(
	ctx context.Context,
	id string,
	name string,
	description string,
	price float64,
	categoryID string,
	stock int32,
	featured bool,
) (database.Product, error) {
	// get user from context
	userId, exists := ctx.Value("userId").(uuid.UUID)
	if !exists {
		s.logger.Error("failed to get user id from context")
		userId = uuid.Nil
	}

	// parsing
	var descriptionValue sql.NullString
	if description != "" {
		descriptionValue.String = description
		descriptionValue.Valid = true
	} else {
		descriptionValue.Valid = false
	}

	var categoryIDValue uuid.NullUUID
	if categoryID != "" {
		categoryIDValue.UUID, _ = uuid.Parse(categoryID)
		categoryIDValue.Valid = true
	} else {
		categoryIDValue.Valid = false
	}

	var stockValue sql.NullInt32
	if stock > 0 {
		stockValue.Int32 = stock
		stockValue.Valid = true
	} else {
		stockValue.Int32 = 0
		stockValue.Valid = true
	}

	var updaterIDValue uuid.NullUUID
	if userId != uuid.Nil {
		updaterIDValue.UUID = userId
		updaterIDValue.Valid = true
	} else {
		updaterIDValue.Valid = false
	}

	productID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid product id", zap.Error(err))
		return database.Product{}, fmt.Errorf("invalid product id: %w", err)
	}

	params := database.UpdateProductParams{
		ID:          productID,
		Name:        name,
		Description: descriptionValue,
		Price:       strconv.FormatFloat(price, 'f', -1, 64),
		CategoryID:  categoryIDValue,
		Stock:       stockValue,
		UpdatedBy:   updaterIDValue,
		Featured:    sql.NullBool{Bool: featured, Valid: true},
	}

	// update product
	updatedProduct, err := s.productRepo.UpdateProduct(ctx, params)
	if err != nil {
		s.logger.Error("failed to update product", zap.Error(err))
		return database.Product{}, err
	}

	return updatedProduct, nil
}

// SoftDeleteProduct marks a product as inactive in the database
func (s *ProductService) SoftDeleteProduct(ctx context.Context, id string) error {
	// get user from context
	productID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid product id", zap.Error(err))
		return fmt.Errorf("invalid product id: %w", err)
	}

	err = s.productRepo.SoftDeleteProduct(ctx, productID)
	if err != nil {
		s.logger.Error("failed to soft delete product", zap.Error(err))
		return fmt.Errorf("failed to soft delete product: %w", err)
	}

	return nil
}

// GetProductsByCategoryID retrieves products by their category ID
func (s *ProductService) GetProductsByCategoryID(ctx context.Context, categoryID string) ([]database.Product, error) {
	// Parse the category ID
	categoryIDValue, err := uuid.Parse(categoryID)

	// Fetch products by category ID from the repository
	products, err := s.productRepo.GetProductsByCategoryID(ctx, categoryIDValue)
	if err != nil {
		s.logger.Error("failed to get products by category ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get products by category ID: %w", err)
	}

	return products, nil
}

// SearchProducts searches for products by name or description
func (s *ProductService) SearchProducts(ctx context.Context, searchTerm string) ([]database.Product, error) {
	// Parse the search term
	searchTermValue := sql.NullString{String: searchTerm, Valid: searchTerm != ""}

	// Fetch products by search term from the repository
	products, err := s.productRepo.SearchProducts(ctx, searchTermValue)
	if err != nil {
		s.logger.Error("failed to search products", zap.Error(err))
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	return products, nil
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
	productID string,
	imageURL string,
) (model.ProductImage, error) {
	// Parse the product ID
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("invalid product ID", zap.Error(err))
		return model.ProductImage{}, fmt.Errorf("invalid product ID: %w", err)
	}

	// Prepare the parameters for creating the product image
	params := database.CreateProductImageParams{
		ProductID: uuid.NullUUID{UUID: productUUID, Valid: true},
		ImageUrl:  imageURL,
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
	id string,
	imageURL string,
) (model.ProductImage, error) {
	// Parse the image ID
	imageUUID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid product image ID", zap.Error(err))
		return model.ProductImage{}, fmt.Errorf("invalid product image ID: %w", err)
	}

	// Prepare the parameters for updating the product image
	params := database.UpdateProductImageParams{
		ID:       imageUUID,
		ImageUrl: imageURL,
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
		CreatedAt: dbImage.CreatedAt.Time,
		UpdatedAt: dbImage.UpdatedAt.Time,
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

// GetProductsByParentCategoryID retrieves products by their parent category ID
func (s *ProductService) GetProductsByParentCategoryID(ctx context.Context, parentCategoryID string) ([]model.Product, error) {
	// Parse the parent category ID
	parentCategoryUUID, err := uuid.Parse(parentCategoryID)

	// Get products by category ID
	var products []model.Product

	productsByCategory, err := s.productRepo.GetProductsByParentCategoryID(ctx, parentCategoryUUID)
	if err != nil {
		s.logger.Error("failed to get products by category ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get products by category ID: %w", err)
	}

	for _, product := range productsByCategory {
		// get the product specifications
		productSpecs, err := s.productSpecificationRepo.ListProductSpecificationsByProductID(ctx, uuid.NullUUID{UUID: product.ID, Valid: true})
		if err != nil {
			s.logger.Error("failed to get product specifications", zap.Error(err))
			return nil, fmt.Errorf("failed to get product specifications: %w", err)
		}

		// get the product variants
		productVariants, err := s.productVariantRepo.ListProductVariantsByProductID(ctx, uuid.NullUUID{UUID: product.ID, Valid: true})
		if err != nil {
			s.logger.Error("failed to get product variants", zap.Error(err))
			return nil, fmt.Errorf("failed to get product variants: %w", err)
		}

		// get the product images
		productImages, err := s.productImageRepo.ListProductImagesByProductID(ctx, uuid.NullUUID{UUID: product.ID, Valid: true})
		if err != nil {
			s.logger.Error("failed to get product images", zap.Error(err))
			return nil, fmt.Errorf("failed to get product images: %w", err)
		}

		// get the product colors
		productColors, err := s.productColorRepo.ListProductColorsByProductID(ctx, uuid.NullUUID{UUID: product.ID, Valid: true})
		if err != nil {
			s.logger.Error("failed to get product colors", zap.Error(err))
			return nil, fmt.Errorf("failed to get product colors: %w", err)
		}

		// map the product to the model
		var specs []model.ProductSpecification
		for _, spec := range productSpecs {
			specs = append(specs, model.ProductSpecification{
				ID:        spec.ID,
				SpecName:  spec.SpecName,
				SpecValue: spec.SpecValue,
			})
		}

		var variants []model.ProductVariant
		for _, variant := range productVariants {
			variants = append(variants, model.ProductVariant{
				ID:           variant.ID,
				VariantName:  variant.VariantName,
				VariantValue: variant.VariantValue,
				Price:        variant.Price,
				Stock:        variant.Stock.Int32,
			})
		}

		var images []model.ProductImage
		for _, image := range productImages {
			images = append(images, model.ProductImage{
				ID:        image.ID,
				ProductID: image.ProductID.UUID.String(),
				S3URL:     s.constructS3URL(image.ImageUrl),
				CreatedAt: image.CreatedAt.Time,
				UpdatedAt: image.UpdatedAt.Time,
			})
		}

		var colors []model.ProductColor
		for _, color := range productColors {
			colors = append(colors, model.ProductColor{
				ID:        color.ID,
				ColorName: color.ColorName,
			})
		}

		products = append(products, model.Product{
			ID:             product.ID,
			Name:           product.Name,
			Description:    product.Description.String,
			Price:          product.Price,
			Stock:          product.Stock.Int32,
			CategoryID:     product.CategoryID.UUID,
			IsActive:       product.IsActive.Bool,
			Featured:       product.Featured.Bool,
			Colors:         colors,
			Specifications: specs,
			Variants:       variants,
			Images:         images,
		})
	}

	return products, nil
}

// CreateProductColor creates a new product color
func (s *ProductService) CreateProductColor(
	ctx context.Context,
	productID string,
	color string,
) (database.ProductColor, error) {
	// Parse the product ID
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("invalid product ID", zap.Error(err))
		return database.ProductColor{}, fmt.Errorf("invalid product ID: %w", err)
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
		return database.ProductColor{}, fmt.Errorf("failed to create product color: %w", err)
	}

	return productColor, nil
}

// ListProductColorsByProductID lists product colors by product ID
func (s *ProductService) ListProductColorsByProductID(
	ctx context.Context,
	productID string,
) ([]database.ProductColor, error) {
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
) (database.ProductColor, error) {
	// Parse the product color ID
	colorUUID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid product color ID", zap.Error(err))
		return database.ProductColor{}, fmt.Errorf("invalid product color ID: %w", err)
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
		return database.ProductColor{}, fmt.Errorf("failed to update product color: %w", err)
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

const (
	// SortByFilterOptions is a list of valid sort by filter options
	SortByFilterOptions = "price_asc,price_desc,name_asc,name_desc"
)

// GetProductsByFilters retrieves products by their parent category ID
func (s *ProductService) GetProductsByFilters(
	ctx context.Context,
	parentCategoryID string,
	subCategories []string,
	colors []string,
	priceFrom string,
	priceTo string,
	sortBy string,
) ([]model.Product, error) {
	parentCategoryUUID, err := uuid.Parse(parentCategoryID)
	if err != nil {
		s.logger.Error("invalid parent category ID", zap.Error(err))
		return nil, fmt.Errorf("invalid parent category ID: %w", err)
	}

	if len(subCategories) == 0 {
		categories, err := s.categoryRepo.GetCategoriesByParentID(ctx, uuid.NullUUID{UUID: parentCategoryUUID, Valid: true})
		if err != nil {
			s.logger.Error("failed to get subcategories", zap.Error(err))
			return nil, fmt.Errorf("failed to get subcategories: %w", err)
		}

		for _, category := range categories {
			subCategories = append(subCategories, category.Name)
		}
	}

	if len(colors) == 0 {
		productColors, err := s.productColorRepo.GetAvailableColorsByParentCategoryID(ctx, parentCategoryUUID)
		if err != nil {
			s.logger.Error("failed to get colors", zap.Error(err))
			return nil, fmt.Errorf("failed to get colors: %w", err)
		}

		for _, color := range productColors {
			colors = append(colors, color)
		}
	}

	if priceFrom == "" {
		priceFrom = "0"
	}

	if priceTo == "" {
		priceTo = "999999"
	}

	priceFromFloat, err := strconv.ParseFloat(priceFrom, 64)
	if err != nil {
		s.logger.Error("invalid price from", zap.Error(err))
		return nil, fmt.Errorf("invalid price from: %w", err)
	}

	priceToFloat, err := strconv.ParseFloat(priceTo, 64)
	if err != nil {
		s.logger.Error("invalid price to", zap.Error(err))
		return nil, fmt.Errorf("invalid price to: %w", err)
	}

	filter := utils.NewFilter()
	for _, subCategory := range subCategories {
		if subCategory != "" {
			filter.Add("ct.name", "=", 1, strings.Trim(subCategory, " "))
		}
	}

	for _, color := range colors {
		if color != "" {
			filter.Add("pc.color_name", "=", 1, strings.Trim(color, " "))
		}
	}

	filter.AddRaw("p.price", fmt.Sprintf("p.price >= %f AND p.price <= %f", priceFromFloat, priceToFloat))

	productsByCategory, err := s.productRepo.GetProductsByFilters(ctx, parentCategoryUUID, filter, sortBy)
	if err != nil {
		s.logger.Error("failed to get products by category ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get products by category ID: %w", err)
	}

	var products []model.Product
	for _, product := range productsByCategory {
		productSpecs, err := s.productSpecificationRepo.ListProductSpecificationsByProductID(ctx, uuid.NullUUID{UUID: product.ID, Valid: true})
		if err != nil {
			s.logger.Error("failed to get product specifications", zap.Error(err))
			return nil, fmt.Errorf("failed to get product specifications: %w", err)
		}

		productVariants, err := s.productVariantRepo.ListProductVariantsByProductID(ctx, uuid.NullUUID{UUID: product.ID, Valid: true})
		if err != nil {
			s.logger.Error("failed to get product variants", zap.Error(err))
			return nil, fmt.Errorf("failed to get product variants: %w", err)
		}

		productImages, err := s.productImageRepo.ListProductImagesByProductID(ctx, uuid.NullUUID{UUID: product.ID, Valid: true})
		if err != nil {
			s.logger.Error("failed to get product images", zap.Error(err))
			return nil, fmt.Errorf("failed to get product images: %w", err)
		}

		productColors, err := s.productColorRepo.ListProductColorsByProductID(ctx, uuid.NullUUID{UUID: product.ID, Valid: true})
		if err != nil {
			s.logger.Error("failed to get product colors", zap.Error(err))
			return nil, fmt.Errorf("failed to get product colors: %w", err)
		}

		var specs []model.ProductSpecification
		for _, spec := range productSpecs {
			specs = append(specs, model.ProductSpecification{
				ID:        spec.ID,
				SpecName:  spec.SpecName,
				SpecValue: spec.SpecValue,
			})
		}

		var variants []model.ProductVariant
		for _, variant := range productVariants {
			variants = append(variants, model.ProductVariant{
				ID:           variant.ID,
				VariantName:  variant.VariantName,
				VariantValue: variant.VariantValue,
				Price:        variant.Price,
				Stock:        variant.Stock.Int32,
			})
		}

		var images []model.ProductImage
		for _, image := range productImages {
			images = append(images, model.ProductImage{
				ID:        image.ID,
				ProductID: image.ProductID.UUID.String(),
				S3URL:     s.constructS3URL(image.ImageUrl),
				CreatedAt: image.CreatedAt.Time,
				UpdatedAt: image.UpdatedAt.Time,
			})
		}

		var colors []model.ProductColor
		for _, color := range productColors {
			colors = append(colors, model.ProductColor{
				ID:        color.ID,
				ColorName: color.ColorName,
			})
		}

		products = append(products, model.Product{
			ID:             product.ID,
			Name:           product.Name,
			Description:    product.Description.String,
			Price:          product.Price,
			Stock:          product.Stock.Int32,
			CategoryID:     product.CategoryID.UUID,
			IsActive:       product.IsActive.Bool,
			Featured:       product.Featured.Bool,
			Colors:         colors,
			Specifications: specs,
			Variants:       variants,
			Images:         images,
		})
	}

	return products, nil
}
