package services

import (
	"context"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/repository"
	"weblineBackend/pkg/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type InquiryService struct {
	productRepo *repository.ProductRepository
	logger      *zap.Logger
	cfg         *appconfig.Config
}

func NewInquiryService(productRepo *repository.ProductRepository, logger *zap.Logger, cfg *appconfig.Config) *InquiryService {
	return &InquiryService{
		productRepo: productRepo,
		logger:      logger,
		cfg:         cfg,
	}
}

func (s *InquiryService) SubmitInquiry(ctx context.Context, productID, email, message string) error {
	// Parse productID to uuid
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("failed to parse productID", zap.Error(err))
		return err
	}

	// Get product
	product, err := s.productRepo.GetProductByID(ctx, productUUID)
	if err != nil {
		s.logger.Error("failed to get product by ID", zap.Error(err))
		return err
	}

	// Send inquiry email
	if err := utils.SendInquiryEmail(s.cfg, product.Name, email, message); err != nil {
		s.logger.Error("failed to send inquiry email", zap.Error(err))
		return err
	}

	return nil
}
