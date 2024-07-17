package services

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
	"weblineBackend/internal/database"
	"weblineBackend/internal/repository"
)

type DiscountService struct {
	logger       *zap.Logger
	discountRepo *repository.DiscountRepository
	productRepo  *repository.ProductRepository
}

func NewDiscountService(logger *zap.Logger, discountRepo *repository.DiscountRepository, productRepo *repository.ProductRepository) *DiscountService {
	return &DiscountService{
		logger:       logger,
		discountRepo: discountRepo,
		productRepo:  productRepo,
	}
}

// CreateDiscount creates a new discount
func (s *DiscountService) CreateDiscount(ctx context.Context, productID uuid.UUID, discount float64, startDate, endDate time.Time) error {
	// check if the product exists
	product, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		s.logger.Error("failed to get product by ID", zap.Error(err))
		return err
	}

	_, err = s.discountRepo.CreateDiscount(ctx, &database.CreateDiscountParams{
		ProductID:          uuid.NullUUID{UUID: product.ID, Valid: true},
		DiscountPercentage: fmt.Sprintf("%f", discount),
		StartDate:          sql.NullTime{Time: startDate, Valid: true},
		EndDate:            sql.NullTime{Time: endDate, Valid: true},
	})
	if err != nil {
		s.logger.Error("failed to create discount", zap.Error(err))
		return err
	}

	return nil
}
