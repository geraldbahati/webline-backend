package services

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProductSizeService struct {
	productSizeRepository *repository.ProductSizeRepository
	logger                *zap.Logger
}

func NewProductSizeService(
	productSizeRepository *repository.ProductSizeRepository,
	logger *zap.Logger,
) *ProductSizeService {
	return &ProductSizeService{productSizeRepository: productSizeRepository, logger: logger}
}

// CreateProductSize creates a new product size
func (s *ProductSizeService) CreateProductSize(ctx context.Context, productID, size string, additionalPrice float64) error {
	// parse product id to uuid
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("failed to parse product id to uuid", zap.Error(err))
		return fmt.Errorf("failed to parse product id to uuid: %w", err)
	}

	// additionalPrice to string
	priceStr := strconv.FormatFloat(additionalPrice, 'f', -1, 64)

	// Prepare parameters for creating new product size
	params := database.CreateProductSizeParams{
		ProductID:       uuid.NullUUID{UUID: productUUID, Valid: true},
		Size:            size,
		AdditionalPrice: sql.NullString{String: priceStr, Valid: true},
	}

	// create new product size
	_, err = s.productSizeRepository.CreateProductSize(ctx, params)
	if err != nil {
		s.logger.Error("failed to create product size", zap.Error(err))
		return fmt.Errorf("failed to create product size: %w", err)
	}

	return nil
}

// ListProductSizesByProductID returns all product sizes by product id
func (s *ProductSizeService) ListProductSizesByProductID(ctx context.Context, productID string) ([]model.ProductSize, error) {
	// parse product id to uuid
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("failed to parse product id to uuid", zap.Error(err))
		return nil, fmt.Errorf("failed to parse product id to uuid: %w", err)
	}

	// get all product sizes by product id
	productSizes, err := s.productSizeRepository.GetProductSizesByProductID(ctx, uuid.NullUUID{UUID: productUUID, Valid: true})
	if err != nil {
		s.logger.Error("failed to get product sizes by product id", zap.Error(err))
		return nil, fmt.Errorf("failed to get product sizes by product id: %w", err)
	}

	var sizes []model.ProductSize
	for _, productSize := range productSizes {
		sizes = append(sizes, model.ProductSize{
			ID:              productSize.ID,
			ProductID:       productSize.ProductID.UUID,
			Size:            productSize.Size,
			AdditionalPrice: productSize.AdditionalPrice.String,
		})
	}

	return sizes, nil
}

// GetProductSizeByID returns product size by id
func (s *ProductSizeService) GetProductSizeByID(ctx context.Context, productSizeID string) (*model.ProductSize, error) {
	// parse product size id to uuid
	productSizeUUID, err := uuid.Parse(productSizeID)
	if err != nil {
		s.logger.Error("failed to parse product size id to uuid", zap.Error(err))
		return nil, fmt.Errorf("failed to parse product size id to uuid: %w", err)
	}

	// get product size by id
	productSize, err := s.productSizeRepository.GetProductSizeByID(ctx, productSizeUUID)
	if err != nil {
		s.logger.Error("failed to get product size by id", zap.Error(err))
		return nil, fmt.Errorf("failed to get product size by id: %w", err)
	}

	return &model.ProductSize{
		ID:              productSize.ID,
		ProductID:       productSize.ProductID.UUID,
		Size:            productSize.Size,
		AdditionalPrice: productSize.AdditionalPrice.String,
	}, nil
}

// UpdateProductSize updates product size by id
func (s *ProductSizeService) UpdateProductSize(ctx context.Context, productSizeID, size string, additionalPrice float64) error {
	// parse product size id to uuid
	productSizeUUID, err := uuid.Parse(productSizeID)
	if err != nil {
		s.logger.Error("failed to parse product size id to uuid", zap.Error(err))
		return fmt.Errorf("failed to parse product size id to uuid: %w", err)
	}

	// additionalPrice to string
	priceStr := strconv.FormatFloat(additionalPrice, 'f', -1, 64)

	// Prepare parameters for updating product size
	params := database.UpdateProductSizeParams{
		ID:              productSizeUUID,
		Size:            size,
		AdditionalPrice: sql.NullString{String: priceStr, Valid: true},
	}

	// update product size
	_, err = s.productSizeRepository.UpdateProductSize(ctx, params)
	if err != nil {
		s.logger.Error("failed to update product size", zap.Error(err))
		return fmt.Errorf("failed to update product size: %w", err)
	}

	return nil
}

// DeleteProductSize deletes product size by id
func (s *ProductSizeService) DeleteProductSize(ctx context.Context, productSizeID string) error {
	// parse product size id to uuid
	productSizeUUID, err := uuid.Parse(productSizeID)
	if err != nil {
		s.logger.Error("failed to parse product size id to uuid", zap.Error(err))
		return fmt.Errorf("failed to parse product size id to uuid: %w", err)
	}

	// delete product size by id
	err = s.productSizeRepository.DeleteProductSize(ctx, productSizeUUID)
	if err != nil {
		s.logger.Error("failed to delete product size", zap.Error(err))
		return fmt.Errorf("failed to delete product size: %w", err)
	}

	return nil
}
