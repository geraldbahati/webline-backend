package services

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
)

type ProductAttributeService struct {
	productAttributeRepo repository.ProductAttributeRepository
	logger               *zap.Logger
}

func NewProductAttributeService(repo repository.ProductAttributeRepository, logger *zap.Logger) *ProductAttributeService {
	return &ProductAttributeService{
		productAttributeRepo: repo,
		logger:               logger,
	}
}

// GetProductAttributesWithValues retrieves the product attributes with values
func (s *ProductAttributeService) GetProductAttributesWithValues(ctx context.Context, categoryID uuid.UUID) (map[string][]model.Attribute, error) {
	attributes, err := s.productAttributeRepo.GetProductAttributesWithValues(ctx, categoryID)
	if err != nil {
		s.logger.Error("failed to get product attributes with values", zap.Error(err))
		return nil, err
	}

	return attributes, nil
}
