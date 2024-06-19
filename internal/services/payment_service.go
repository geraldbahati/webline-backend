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

type PaymentService struct {
	paymentRepository *repository.PaymentRepository
	logger            *zap.Logger
}

func NewPaymentService(pr *repository.PaymentRepository, logger *zap.Logger) *PaymentService {
	return &PaymentService{
		paymentRepository: pr,
		logger:            logger,
	}
}

// CreatePayment creates a new payment
func (s *PaymentService) CreatePayment(ctx context.Context, orderID string, amount float64, paymentMethodID int) (*model.OrderPayment, error) {
	var orderUUID uuid.NullUUID

	if orderID == "" {
		s.logger.Error("OrderID is required")
		return nil, fmt.Errorf("orderID is required")
	} else {
		id, err := uuid.Parse(orderID)
		if err != nil {
			s.logger.Error("Failed to parse orderID", zap.Error(err), zap.String("orderID", orderID))
			return nil, fmt.Errorf("parse orderID: %w", err)
		}
		orderUUID = uuid.NullUUID{UUID: id, Valid: true}
	}

	payment := &database.CreatePaymentParams{
		OrderID:         orderUUID,
		Amount:          strconv.FormatFloat(amount, 'f', -1, 64),
		PaymentMethodID: sql.NullInt32{Int32: int32(paymentMethodID), Valid: true},
	}

	orderPayment, err := s.paymentRepository.CreatePayment(ctx, payment)
	if err != nil {
		s.logger.Error("Failed to create payment", zap.Error(err), zap.Any("paymentParams", payment))
		return nil, err
	}

	return orderPayment, nil

}
