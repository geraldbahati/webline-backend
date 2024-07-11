package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"log"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
)

type ProductSEOService struct {
	logger      *zap.Logger
	config      *appconfig.Config
	productRepo *repository.ProductRepository
}

func NewProductSEOService(logger *zap.Logger, config *appconfig.Config, productRepo *repository.ProductRepository) *ProductSEOService {
	return &ProductSEOService{
		logger:      logger,
		config:      config,
		productRepo: productRepo,
	}
}

// UpdateProductSEO updates the SEO information of a product
func (s *ProductSEOService) UpdateProductSEO(ctx context.Context, productID *uuid.UUID, partNumber, seoTitle, seoDescription, seoKeywords string) error {
	// Create the params
	params := &database.UpdateProductSEOParams{
		ID: *productID,
		PartNumber: sql.NullString{
			String: partNumber,
			Valid:  partNumber != "",
		},
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

	return nil
}

// GetProductSEO returns the SEO information of a product
func (s *ProductSEOService) GetProductSEO(ctx context.Context, productID *uuid.UUID) (*model.ProductSEO, error) {
	// Get the product SEO
	productSEO, err := s.productRepo.GetProductSEO(ctx, *productID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, sql.ErrNoRows
		default:
			s.logger.Error("failed to get product SEO", zap.Error(err))
			return nil, err
		}
	}

	log.Printf("AWSBucketName: %s AWSRegion: %s", s.config.AWSBucketName, s.config.AWSRegion)

	if productSEO.ImageUrl != "" {
		imageUrl := s.constructS3URL(productSEO.ImageUrl)
		productSEO.ImageUrl = imageUrl
	}

	return productSEO, nil
}

// constructS3URL constructs the S3 URL for a given file path
func (s *ProductSEOService) constructS3URL(filePath string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.AWSBucketName, s.config.AWSRegion, filePath)
}
