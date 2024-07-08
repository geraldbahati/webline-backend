package services

import (
	"go.uber.org/zap"
	"weblineBackend/internal/repository"
)

type DiscountService struct {
	logger       *zap.Logger
	discountRepo *repository.DiscountRepository
}

func NewDiscountService(logger *zap.Logger, discountRepo *repository.DiscountRepository) *DiscountService {
	return &DiscountService{
		logger:       logger,
		discountRepo: discountRepo,
	}
}
