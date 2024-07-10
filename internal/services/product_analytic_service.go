package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"strconv"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
)

type ProductAnalyticService struct {
	logger           *zap.Logger
	config           *appconfig.Config
	analyticRep      *repository.ProductAnalyticRepository
	productImageRepo *repository.ProductImageRepository
	discountRepo     *repository.DiscountRepository
}

func NewProductAnalyticService(logger *zap.Logger, config *appconfig.Config, analyticRep *repository.ProductAnalyticRepository, productImageRepo *repository.ProductImageRepository, discountRepo *repository.DiscountRepository) *ProductAnalyticService {
	return &ProductAnalyticService{
		logger:           logger,
		config:           config,
		analyticRep:      analyticRep,
		productImageRepo: productImageRepo,
		discountRepo:     discountRepo,
	}
}

func (s *ProductAnalyticService) getProductImages(ctx context.Context, productID uuid.UUID) ([]model.ProductImage, error) {
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
			CreatedAt: image.CreatedAt.Time,
			UpdatedAt: image.UpdatedAt.Time,
		})
	}

	return images, nil
}

func (s *ProductAnalyticService) getProductDiscountPercentage(ctx context.Context, productID uuid.UUID) (float64, error) {
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

// constructS3URL constructs the S3 URL for a given file path
func (s *ProductAnalyticService) constructS3URL(filePath string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.AWSBucketName, s.config.AWSRegion, filePath)
}

// GetBestSellerProducts returns the best seller products
func (s *ProductAnalyticService) GetBestSellerProducts(ctx context.Context, limit int32) ([]*model.Product, error) {
	// get best seller products from repository
	products, err := s.analyticRep.GetBestSellerProducts(ctx, limit)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			s.logger.Info("No best seller products found")
			return nil, err
		default:
			s.logger.Error("Failed to get best seller products", zap.Error(err))
			return nil, err
		}
	}

	for _, product := range products {
		// get image url for the product
		productImages, err := s.getProductImages(ctx, product.ID)
		if err != nil {
			return nil, err
		}

		// get discount percentage for the product
		discountPercent, err := s.getProductDiscountPercentage(ctx, product.ID)
		if err != nil {
			return nil, err
		}

		// update product with images and discount percentage
		product.ImageURL = productImages[0].S3URL
		product.DiscountPercent = discountPercent
	}
	return products, nil
}

// GetFeaturedProducts returns the featured products
func (s *ProductAnalyticService) GetFeaturedProducts(ctx context.Context, limit int32) ([]*model.Product, error) {
	// get featured products from repository
	products, err := s.analyticRep.GetFeaturedProducts(ctx, limit)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			s.logger.Info("No featured products found")
			return nil, err
		default:
			s.logger.Error("Failed to get featured products", zap.Error(err))
			return nil, err
		}
	}

	for _, product := range products {
		// get image url for the product
		productImages, err := s.getProductImages(ctx, product.ID)
		if err != nil {
			return nil, err
		}

		// get discount percentage for the product
		discountPercent, err := s.getProductDiscountPercentage(ctx, product.ID)
		if err != nil {
			return nil, err
		}

		// update product with images and discount percentage
		product.ImageURL = productImages[0].S3URL
		product.DiscountPercent = discountPercent
	}
	return products, nil
}

// GetNewArrivalProducts returns the new arrival products
func (s *ProductAnalyticService) GetNewArrivalProducts(ctx context.Context, limit int32) ([]*model.Product, error) {
	// get new arrival products from repository
	products, err := s.analyticRep.GetNewArrivalProducts(ctx, limit)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			s.logger.Info("No new arrival products found")
			return nil, err
		default:
			s.logger.Error("Failed to get new arrival products", zap.Error(err))
			return nil, err
		}
	}

	for _, product := range products {
		// get image url for the product
		productImages, err := s.getProductImages(ctx, product.ID)
		if err != nil {
			return nil, err
		}

		// get discount percentage for the product
		discountPercent, err := s.getProductDiscountPercentage(ctx, product.ID)
		if err != nil {
			return nil, err
		}

		// update product with images and discount percentage
		product.ImageURL = productImages[0].S3URL
		product.DiscountPercent = discountPercent
	}
	return products, nil
}
