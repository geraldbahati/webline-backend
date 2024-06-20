package services

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
	"weblineBackend/pkg/mpesa"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PaymentService struct {
	paymentRepository *repository.PaymentRepository
	orderRepository   *repository.OrderRepository
	logger            *zap.Logger
	cfg               *appconfig.Config
}

func NewPaymentService(pr *repository.PaymentRepository, o *repository.OrderRepository, logger *zap.Logger, cfg *appconfig.Config) *PaymentService {
	return &PaymentService{
		paymentRepository: pr,
		orderRepository:   o,
		logger:            logger,
		cfg:               cfg,
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

// PayOrderWithMpesa simulates a payment with Mpesa
func (s *PaymentService) PayOrderWithMpesa(ctx context.Context, orderID, phone string) (string, error) {
	// Get order
	// Parse orderID to uuid
	orderUUID, err := uuid.Parse(orderID)
	if err != nil {
		s.logger.Error("Failed to parse orderID", zap.Error(err), zap.String("orderID", orderID))
		return "", fmt.Errorf("parse orderID: %w", err)
	}

	// Get payment by orderID
	payment, err := s.paymentRepository.GetPaymentByOrderID(ctx, order.ID)
	if err != nil {
		s.logger.Error("Failed to get payment", zap.Error(err), zap.String("orderID", orderID))
		return "", fmt.Errorf("get payment: %w", err)
	}

	// Pay order
	token, err := mpesa.GetOAuthToken(s.cfg.ConsumerKey, s.cfg.ConsumerSecret)
	if err != nil {
		s.logger.Error("Failed to get OAuth token", zap.Error(err))
		return "", fmt.Errorf("get OAuth token: %w", err)
	}

	// Generate password
	timestamp := time.Now().Format("20060102150405")
	password := mpesa.GeneratePassword(s.cfg.BusinessShortCode, s.cfg.Passkey, timestamp)
	amount
}
