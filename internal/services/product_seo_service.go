package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProductSEOService struct {
	logger      *zap.Logger
	config      *appconfig.Config
	productRepo *repository.ProductRepository
	cache       CacheService
}

func NewProductSEOService(logger *zap.Logger, config *appconfig.Config, productRepo *repository.ProductRepository, cache CacheService) *ProductSEOService {
	return &ProductSEOService{
		logger:      logger,
		config:      config,
		productRepo: productRepo,
		cache:       cache,
	}
}

// UpdateProductSEO updates the SEO information of a product
func (s *ProductSEOService) UpdateProductSEO(ctx context.Context, productID *uuid.UUID, partNumber, seoTitle, seoDescription, seoKeywords string) error {
	// Create the params
	params := &database.UpdateProductSEOParams{
		ID:         *productID,
		PartNumber: partNumber,
		MetaTitle: sql.NullString{
			String: seoTitle,
			Valid:  seoTitle != "",
		},
		MetaDescription: sql.NullString{
			String: seoDescription,
			Valid:  seoDescription != "",
		},
		MetaKeywords: sql.NullString{
			String: seoKeywords,
			Valid:  seoKeywords != "",
		},
	}

	// Update the product SEO
	err := s.productRepo.UpdateProductSEO(ctx, params)
	if err != nil {
		s.logger.Error("failed to update product SEO", zap.Error(err))
		return fmt.Errorf("failed to update product SEO: %w", err)
	}

	// Invalidate the cache entry
	slug, err := s.productRepo.GetProductSlugByProductID(ctx, *productID)
	if err != nil {
		s.logger.Error("failed to get product slug by ID", zap.Error(err))
		return fmt.Errorf("failed to get product slug by ID: %w", err)
	}

	cacheKey := ProductSEOKey(slug)
	err = s.cache.Delete(ctx, cacheKey)
	if err != nil {
		s.logger.Error("failed to delete cache entry", zap.Error(err))
	}

	return nil
}

// GetProductSEO returns the SEO information of a product
func (s *ProductSEOService) GetProductSEO(ctx context.Context, slug string) (*model.ProductSEO, error) {
	cacheKey := ProductSEOKey(slug)

	var productSEO model.ProductSEO

	// Use GetOrSet to handle cache retrieval and population
	err := s.cache.GetOrSet(ctx, cacheKey, &productSEO, func() error {
		// Fetch from repository if cache miss
		fetchedSEO, err := s.productRepo.GetProductSEO(ctx, slug)
		if err != nil {
			return err
		}

		// Transform fetched data if necessary
		if fetchedSEO.ImageUrl != "" {
			fetchedSEO.ImageUrl = s.constructS3URL(fetchedSEO.ImageUrl)
		}

		// Assign to destination
		productSEO = *fetchedSEO

		return nil
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		// Log and return other errors
		s.logger.Error("failed to get product SEO", zap.Error(err))
		return nil, err
	}

	return &productSEO, nil
}

// constructS3URL constructs the S3 URL for a given file path
func (s *ProductSEOService) constructS3URL(filePath string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.AWSBucketName, s.config.AWSRegion, filePath)
}
