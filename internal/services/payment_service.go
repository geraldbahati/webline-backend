package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
	paymentRepository   *repository.PaymentRepository
	orderRepository     *repository.OrderRepository
	orderItemRepository *repository.OrderItemRepository
	logger              *zap.Logger
	cfg                 *appconfig.Config
}

func NewPaymentService(pr *repository.PaymentRepository, o *repository.OrderRepository, oi *repository.OrderItemRepository, logger *zap.Logger, cfg *appconfig.Config) *PaymentService {
	return &PaymentService{
		paymentRepository:   pr,
		orderRepository:     o,
		orderItemRepository: oi,
		logger:              logger,
		cfg:                 cfg,
	}
}

// CreatePayment creates a new payment
func (s *PaymentService) CreatePayment(ctx context.Context, orderID string, amount float64, paymentMethodID int) (*model.OrderPayment, error) {
	var orderUUID uuid.UUID

	if orderID == "" {
		s.logger.Error("OrderID is required")
		return nil, fmt.Errorf("orderID is required")
	} else {
		id, err := uuid.Parse(orderID)
		if err != nil {
			s.logger.Error("Failed to parse orderID", zap.Error(err), zap.String("orderID", orderID))
			return nil, fmt.Errorf("parse orderID: %w", err)
		}
		orderUUID = id
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
func (s *PaymentService) PayOrderWithMpesa(ctx context.Context, orderID, phone string) error {
	businessShortCode := s.cfg.BusinessShortCode
	passkey := s.cfg.Passkey
	callbackURL := s.cfg.CallbackURL
	consumerKey := s.cfg.ConsumerKey
	consumerSecret := s.cfg.ConsumerSecret
	accountReference := s.cfg.AccountReference

	// Get order
	// Parse orderID to uuid
	orderUUID, err := uuid.Parse(orderID)
	if err != nil {
		s.logger.Error("Failed to parse orderID", zap.Error(err), zap.String("orderID", orderID))
		return fmt.Errorf("parse orderID: %w", err)
	}

	// Get payment by orderID
	payment, err := s.paymentRepository.GetPaymentByOrderID(ctx, orderUUID)
	if err != nil {
		s.logger.Error("Failed to get payment", zap.Error(err), zap.String("orderID", orderID))
		return fmt.Errorf("get payment: %w", err)
	}

	// Pay order
	token, err := mpesa.GetOAuthToken(consumerKey, consumerSecret)
	if err != nil {
		s.logger.Error("Failed to get OAuth token", zap.Error(err))
		return fmt.Errorf("get OAuth token: %w", err)
	}

	s.logger.Info("OAuth token", zap.Any("token", token))

	// Generate password
	timestamp := time.Now().Format("20060102150405")
	password := mpesa.GeneratePassword(businessShortCode, passkey, timestamp)
	paymentDescription := "Payment for order " + orderUUID.String()

	// amount
	amountFloat, err := strconv.ParseFloat(payment.Amount, 64)
	if err != nil {
		s.logger.Error("Failed to parse amount", zap.Error(err))
		return fmt.Errorf("parse amount: %w", err)
	}

	log.Println("Amount: ", amountFloat)

	mpesaResp, err := mpesa.InitiateMpesaPayment(
		token.AccessToken,
		businessShortCode,
		password,
		timestamp,
		fmt.Sprintf("%.0f", amountFloat),
		phone,
		callbackURL,
		accountReference,
		paymentDescription,
	)
	if err != nil {
		s.logger.Error("Failed to initiate mpesa payment", zap.Error(err))
		return fmt.Errorf("initiate mpesa payment: %w", err)
	}

	s.logger.Info("Mpesa payment initiated", zap.Any("mpesaResp", mpesaResp))

	// Update payment id
	err = s.paymentRepository.UpdateCheckoutRequestID(ctx, orderUUID, mpesaResp.CheckoutRequestID)
	if err != nil {
		s.logger.Error("Failed to update payment with Mpesa CheckoutRequestID", zap.Error(err))
		return fmt.Errorf("failed to update payment with Mpesa CheckoutRequestID %w", err)
	}

	return nil
}

func (s *PaymentService) ProcessMpesaCallback(ctx context.Context, callbackResponse mpesa.MpesaCallbackResponse) error {
	s.logger.Info("Processing Mpesa callback", zap.Any("callbackResponse", callbackResponse))
	checkoutRequestID := callbackResponse.Body.StkCallback.CheckoutRequestID
	resultCode := callbackResponse.Body.StkCallback.ResultCode
	resultDesc := callbackResponse.Body.StkCallback.ResultDesc

	// Determine status based on resultCode
	status := 3
	if resultCode == 0 {
		status = 2 // Successful transaction
	}

	var amount float64
	if callbackResponse.Body.StkCallback.CallbackMetadata.Item != nil {
		for _, item := range callbackResponse.Body.StkCallback.CallbackMetadata.Item {
			if item.Name == "Amount" {
				switch v := item.Value.(type) {
				case string:
					var err error
					amount, err = strconv.ParseFloat(v, 64)
					if err != nil {
						s.logger.Error("Failed to parse amount", zap.Error(err), zap.String("amount", v))
						return fmt.Errorf("parse amount: %w", err)
					}
				case float64:
					amount = v
				default:
					s.logger.Error("Unexpected type for amount", zap.Any("type", v))
					return fmt.Errorf("unexpected type for amount: %T", v)
				}
			}
		}
	} else {
		s.logger.Warn("CallbackMetadata.Item is nil")
	}

	// Update payment status in the database
	err := s.paymentRepository.UpdatePaymentStatus(ctx, database.UpdatePaymentStatusParams{
		CheckoutRequestID: checkoutRequestID,
		PaymentStatusID:   sql.NullInt32{Int32: int32(status), Valid: true},
		Amount:            strconv.FormatFloat(amount, 'f', -1, 64),
		ResultCode:        sql.NullInt32{Int32: int32(resultCode), Valid: true},
		ResultDesc:        sql.NullString{String: resultDesc, Valid: true},
	})
	if err != nil {
		s.logger.Error("Failed to update payment status", zap.Error(err), zap.String("CheckoutRequestID", checkoutRequestID))
		return fmt.Errorf("update payment status: %w", err)
	}

	if status == 3 {
		s.logger.Error("Mpesa transaction failed", zap.String("ResultDesc", resultDesc))
		return fmt.Errorf("mpesa transaction failed: %s", resultDesc)
	}

	return nil
}

// GetPaymentStatus gets the payment status
func (s *PaymentService) GetPaymentStatus(ctx context.Context, orderID string) (string, error) {
	// Parse orderID to uuid
	orderUUID, err := uuid.Parse(orderID)
	if err != nil {
		s.logger.Error("Failed to parse orderID", zap.Error(err), zap.String("orderID", orderID))
		return "failed", fmt.Errorf("parse orderID: %w", err)
	}

	payment, err := s.paymentRepository.GetPaymentByOrderID(ctx, orderUUID)
	if err != nil {
		s.logger.Error("Failed to get payment", zap.Error(err), zap.String("orderID", orderID))
		return "failed", fmt.Errorf("get payment: %w", err)
	}

	status, err := s.paymentRepository.GetStatusByID(ctx, payment.PaymentStatusID)
	if err != nil {
		s.logger.Error("Failed to get payment status", zap.Error(err), zap.Int32("paymentStatusID", payment.PaymentStatusID))
		return "failed", fmt.Errorf("get payment status: %w", err)
	}
	return status, nil
}
