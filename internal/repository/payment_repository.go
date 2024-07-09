package repository

import (
	"context"
	"database/sql"
	"fmt"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PaymentRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewPaymentRepository(db *sql.DB, logger *zap.Logger) *PaymentRepository {
	return &PaymentRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *PaymentRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	q := database.New(tx)
	if err := fn(q); err != nil {
		r.logger.Error("transaction failed, rolling back", zap.Error(err))
		if rbErr := tx.Rollback(); rbErr != nil {
			r.logger.Error("rollback failed", zap.Error(rbErr))
			return fmt.Errorf("rollback transaction: %w", rbErr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		r.logger.Error("commit failed", zap.Error(err))
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// CreatePayment creates a new payment
func (r *PaymentRepository) CreatePayment(ctx context.Context, payment *database.CreatePaymentParams) (*model.OrderPayment, error) {
	var orderPayment *model.OrderPayment
	if err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		orderPay, err := q.CreatePayment(ctx, *payment)
		if err != nil {
			r.logger.Error("Failed to create payment", zap.Error(err), zap.Any("paymentParams", payment))
			return fmt.Errorf("create payment: %w", err)
		}

		orderPayment = &model.OrderPayment{
			ID:                orderPay.ID,
			OrderID:           orderPay.OrderID,
			CheckoutRequestID: orderPay.CheckoutRequestID,
			Amount:            orderPay.Amount,
			CreatedAt:         orderPay.CreatedAt.Time,
			PaymentMethodID:   orderPay.PaymentMethodID.Int32,
			PaymentStatusID:   orderPay.PaymentStatusID.Int32,
		}

		return nil
	}); err != nil {
		r.logger.Error("Failed to create payment", zap.Error(err), zap.Any("paymentParams", payment))
		return nil, err
	}

	r.logger.Info("Payment created successfully", zap.String("checkoutRequestID", orderPayment.CheckoutRequestID))
	return orderPayment, nil
}

// UpdatePaymentStatus updates the payment status
func (r *PaymentRepository) UpdatePaymentStatus(ctx context.Context, params database.UpdatePaymentStatusParams) error {
	if err := r.execTx(ctx, func(q *database.Queries) error {
		if err := q.UpdatePaymentStatus(ctx, params); err != nil {
			r.logger.Error("Failed to update payment status", zap.Error(err), zap.String("checkoutRequestID", params.CheckoutRequestID), zap.Int32("statusID", params.PaymentStatusID.Int32))
			return fmt.Errorf("update payment status: %w", err)
		}
		return nil
	}); err != nil {
		r.logger.Error("Failed to update payment status", zap.Error(err), zap.String("checkoutRequestID", params.CheckoutRequestID), zap.Int32("statusID", params.PaymentStatusID.Int32))
		return err
	}
	return nil
}

// GetPaymentsByOrderID returns payments by order ID
func (r *PaymentRepository) GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*model.OrderPayment, error) {
	payment, err := r.Queries.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		r.logger.Error("Failed to get payments by order ID", zap.Error(err), zap.String("orderID", orderID.String()))
		return nil, fmt.Errorf("get payments by order ID: %w", err)
	}

	orderPayment := model.OrderPayment{
		ID:                payment.ID,
		OrderID:           payment.OrderID,
		CheckoutRequestID: payment.CheckoutRequestID,
		Amount:            payment.Amount,
		CreatedAt:         payment.CreatedAt.Time,
		PaymentMethodID:   payment.PaymentMethodID.Int32,
		PaymentStatusID:   payment.PaymentStatusID.Int32,
		ResultCode:        payment.ResultCode.Int32,
		ResultDesc:        payment.ResultDesc.String,
	}

	return &orderPayment, nil

}

// GetAllPayments returns all payments
func (r *PaymentRepository) GetAllPayments(ctx context.Context) ([]model.OrderPayment, error) {
	payments, err := r.Queries.GetAllPayments(ctx)
	if err != nil {
		r.logger.Error("Failed to get all payments", zap.Error(err))
		return nil, fmt.Errorf("get all payments: %w", err)
	}

	var orderPayments []model.OrderPayment
	for _, payment := range payments {
		orderPayments = append(orderPayments, model.OrderPayment{
			ID:                payment.ID,
			OrderID:           payment.OrderID,
			CheckoutRequestID: payment.CheckoutRequestID,
			Amount:            payment.Amount,
			CreatedAt:         payment.CreatedAt.Time,
			PaymentMethodID:   payment.PaymentMethodID.Int32,
			PaymentStatusID:   payment.PaymentStatusID.Int32,
			ResultCode:        payment.ResultCode.Int32,
			ResultDesc:        payment.ResultDesc.String,
		})
	}

	return orderPayments, nil
}

// UpdateCheckoutRequestID updates the checkout request ID
func (r *PaymentRepository) UpdateCheckoutRequestID(ctx context.Context, orderID uuid.UUID, checkoutRequestID string) error {
	if err := r.execTx(ctx, func(q *database.Queries) error {
		if err := q.UpdateCheckoutRequestIDByOrderID(ctx, database.UpdateCheckoutRequestIDByOrderIDParams{
			OrderID:           orderID,
			CheckoutRequestID: checkoutRequestID,
		}); err != nil {
			r.logger.Error("Failed to update checkout request ID", zap.Error(err), zap.String("checkoutRequestID", checkoutRequestID))
			return fmt.Errorf("update checkout request ID: %w", err)
		}
		return nil
	}); err != nil {
		r.logger.Error("Failed to update checkout request ID", zap.Error(err), zap.String("checkoutRequestID", checkoutRequestID))
		return err
	}
	return nil
}

// GetStatusByID returns payment status by ID
func (r *PaymentRepository) GetStatusByID(ctx context.Context, statusID int32) (string, error) {
	status, err := r.Queries.GetStatusByID(ctx, statusID)
	if err != nil {
		r.logger.Error("Failed to get status by ID", zap.Error(err), zap.Int32("statusID", statusID))
		return "", fmt.Errorf("get status by ID: %w", err)
	}

	return status, nil
}

// ChangeOrderPaymentMethod changes the payment method of an order
func (r *PaymentRepository) ChangeOrderPaymentMethod(ctx context.Context, orderID uuid.UUID, paymentMethod string) (string, error) {
	var orderNumber string
	if err := r.execTx(ctx, func(q *database.Queries) error {
		orderNum, err := q.ChangeOrderPaymentMethod(ctx, database.ChangeOrderPaymentMethodParams{
			OrderID: orderID,
			Method:  paymentMethod,
		})
		if err != nil {
			r.logger.Error("Failed to change payment method", zap.Error(err), zap.String("orderID", orderID.String()), zap.String("paymentMethod", paymentMethod))
			return fmt.Errorf("change payment method: %w", err)
		}
		orderNumber = orderNum.String
		return nil
	}); err != nil {
		r.logger.Error("Failed to change payment method", zap.Error(err), zap.String("orderID", orderID.String()), zap.String("paymentMethod", paymentMethod))
		return "", fmt.Errorf("change payment method: %w", err)
	}

	r.logger.Info("Payment method changed successfully", zap.String("orderID", orderID.String()), zap.String("paymentMethod", paymentMethod))
	return orderNumber, nil
}
