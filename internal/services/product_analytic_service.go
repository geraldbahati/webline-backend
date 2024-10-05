package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
	"weblineBackend/pkg/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProductAnalyticService struct {
	logger           *zap.Logger
	config           *appconfig.Config
	analyticRep      *repository.ProductAnalyticRepository
	productImageRepo *repository.ProductImageRepository
	discountRepo     *repository.DiscountRepository
	cacheService     CacheService
}

// NewProductAnalyticService creates a new instance of ProductAnalyticService.
func NewProductAnalyticService(
	logger *zap.Logger,
	config *appconfig.Config,
	analyticRep *repository.ProductAnalyticRepository,
	productImageRepo *repository.ProductImageRepository,
	discountRepo *repository.DiscountRepository,
	cacheService CacheService,
) *ProductAnalyticService {
	return &ProductAnalyticService{
		logger:           logger,
		config:           config,
		analyticRep:      analyticRep,
		productImageRepo: productImageRepo,
		discountRepo:     discountRepo,
		cacheService:     cacheService,
	}
}

// getProductImages retrieves product images and constructs their S3 URLs.
func (s *ProductAnalyticService) getProductImages(ctx context.Context, productID uuid.UUID) ([]model.ProductImage, error) {
	productImages, err := s.productImageRepo.ListProductImagesByProductID(ctx, uuid.NullUUID{UUID: productID, Valid: true})
	if err != nil {
		s.logger.Error("Failed to get product images", zap.Error(err))
		return nil, fmt.Errorf("get product images: %w", err)
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

// getProductDiscountPercentage retrieves the discount percentage for a product.
func (s *ProductAnalyticService) getProductDiscountPercentage(ctx context.Context, productID uuid.UUID) (float64, error) {
	discount, err := s.discountRepo.GetDiscountByProductID(ctx, &productID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Error("Failed to get product discount", zap.Error(err))
		return 0, fmt.Errorf("get product discount: %w", err)
	}

	if discount != nil {
		discountPercentage, err := strconv.ParseFloat(discount.DiscountPercentage, 64)
		if err != nil {
			s.logger.Error("Failed to parse discount percentage", zap.Error(err))
			return 0, fmt.Errorf("parse discount percentage: %w", err)
		}
		return discountPercentage, nil
	}

	// No discount found; return 0%
	return 0, nil
}

// constructS3URL constructs the S3 URL for a given file path.
func (s *ProductAnalyticService) constructS3URL(filePath string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.AWSBucketName, s.config.AWSRegion, filePath)
}

// GetBestSellerProducts returns the best seller products with caching.
func (s *ProductAnalyticService) GetBestSellerProducts(ctx context.Context, limit int32) ([]*model.Product, error) {
	// Use standardized cache key function
	cacheKey := ProductBestSellersKey(limit)

	var products []*model.Product

	err := s.cacheService.GetOrSet(ctx, cacheKey, &products, func() error {
		// Fetch from repository
		fetchedProducts, err := s.analyticRep.GetBestSellerProducts(ctx, limit)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.logger.Info("No best seller products found")
				return nil
			}
			s.logger.Error("Failed to get best seller products", zap.Error(err))
			return err
		}

		// Update products with images and discounts
		for _, product := range fetchedProducts {
			productImages, err := s.getProductImages(ctx, product.ID)
			if err != nil {
				return err
			}

			discountPercent, err := s.getProductDiscountPercentage(ctx, product.ID)
			if err != nil {
				return err
			}

			product.Price = utils.RoundPriceString(product.Price)
			if len(productImages) > 0 {
				product.ImageURL = productImages[0].S3URL
			}
			product.DiscountPercent = discountPercent
		}

		// Assign to products
		products = fetchedProducts
		return nil
	})

	if err != nil {
		s.logger.Error("Failed to get best seller products", zap.Error(err))
		return nil, err
	}

	s.logger.Debug("Best seller products retrieved", zap.Int("count", len(products)))
	return products, nil
}

// GetFeaturedProducts returns the featured products with caching.
func (s *ProductAnalyticService) GetFeaturedProducts(ctx context.Context, limit int32) ([]*model.Product, error) {
	cacheKey := ProductFeaturedKey(limit)

	var products []*model.Product

	err := s.cacheService.GetOrSet(ctx, cacheKey, &products, func() error {
		// Fetch from repository
		fetchedProducts, err := s.analyticRep.GetFeaturedProducts(ctx, limit)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.logger.Info("No featured products found")
				return nil
			}
			s.logger.Error("Failed to get featured products", zap.Error(err))
			return err
		}

		// Update products with images and discounts
		for _, product := range fetchedProducts {
			productImages, err := s.getProductImages(ctx, product.ID)
			if err != nil {
				return err
			}

			discountPercent, err := s.getProductDiscountPercentage(ctx, product.ID)
			if err != nil {
				return err
			}

			product.Price = utils.RoundPriceString(product.Price)
			if len(productImages) > 0 {
				product.ImageURL = productImages[0].S3URL
			}
			product.DiscountPercent = discountPercent
		}

		// Assign to products
		products = fetchedProducts
		return nil
	})

	if err != nil {
		s.logger.Error("Failed to get featured products", zap.Error(err))
		return nil, err
	}

	s.logger.Debug("Featured products retrieved", zap.Int("count", len(products)))
	return products, nil
}

// GetNewArrivalProducts returns the new arrival products with caching.
func (s *ProductAnalyticService) GetNewArrivalProducts(ctx context.Context, limit int32) ([]*model.Product, error) {
	cacheKey := ProductNewArrivalsKey(limit)

	var products []*model.Product

	err := s.cacheService.GetOrSet(ctx, cacheKey, &products, func() error {
		// Fetch from repository
		fetchedProducts, err := s.analyticRep.GetNewArrivalProducts(ctx, limit)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.logger.Info("No new arrival products found")
				return nil
			}
			s.logger.Error("Failed to get new arrival products", zap.Error(err))
			return err
		}

		// Update products with images and discounts
		for _, product := range fetchedProducts {
			productImages, err := s.getProductImages(ctx, product.ID)
			if err != nil {
				return err
			}

			discountPercent, err := s.getProductDiscountPercentage(ctx, product.ID)
			if err != nil {
				return err
			}

			product.Price = utils.RoundPriceString(product.Price)
			if len(productImages) > 0 {
				product.ImageURL = productImages[0].S3URL
			}
			product.DiscountPercent = discountPercent
		}

		// Assign to products
		products = fetchedProducts
		return nil
	})

	if err != nil {
		s.logger.Error("Failed to get new arrival products", zap.Error(err))
		return nil, err
	}

	s.logger.Debug("New arrival products retrieved", zap.Int("count", len(products)))
	return products, nil
}

// GetDailyDealsProducts returns the daily deals products with caching.
func (s *ProductAnalyticService) GetDailyDealsProducts(ctx context.Context) ([]*model.Product, error) {
	cacheKey := ProductDailyDealsKey()

	var products []*model.Product

	err := s.cacheService.GetOrSet(ctx, cacheKey, &products, func() error {
		// Fetch from repository
		fetchedProducts, err := s.analyticRep.GetDailyDealsProducts(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.logger.Info("No daily deals products found")
				return nil
			}
			s.logger.Error("Failed to get daily deals products", zap.Error(err))
			return err
		}

		// Update products with S3 URLs and rounding price
		for _, product := range fetchedProducts {
			product.Price = utils.RoundPriceString(product.Price)
			if product.ImageURL != "" {
				product.ImageURL = s.constructS3URL(product.ImageURL)
			}
		}

		// Assign to products
		products = fetchedProducts
		return nil
	})

	if err != nil {
		s.logger.Error("Failed to get daily deals products", zap.Error(err))
		return nil, err
	}

	s.logger.Debug("Daily deals products retrieved", zap.Int("count", len(products)))
	return products, nil
}
