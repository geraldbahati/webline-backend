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
			ID:              orderPay.ID,
			OrderID:         orderPay.OrderID.UUID,
			PaymentID:       orderPay.PaymentID,
			Amount:          orderPay.Amount,
			CreatedAt:       orderPay.CreatedAt.Time,
			PaymentMethodID: orderPay.PaymentMethodID.Int32,
			PaymentStatusID: orderPay.PaymentStatusID.Int32,
		}

		return nil
	}); err != nil {
		r.logger.Error("Failed to create payment", zap.Error(err), zap.Any("paymentParams", payment))
		return nil, err
	}

	r.logger.Info("Payment created successfully", zap.String("paymentID", orderPayment.PaymentID))
	return orderPayment, nil
}

// UpdatePaymentStatus updates the payment status
func (r *PaymentRepository) UpdatePaymentStatus(ctx context.Context, paymentID string, statusID int32) error {
	if err := r.execTx(ctx, func(q *database.Queries) error {
		if err := q.UpdatePaymentStatus(ctx, database.UpdatePaymentStatusParams{
			PaymentID: paymentID,
			PaymentStatusID: sql.NullInt32{
				Int32: statusID,
				Valid: true,
			},
		}); err != nil {
			r.logger.Error("Failed to update payment status", zap.Error(err), zap.String("paymentID", paymentID), zap.Int32("statusID", statusID))
			return fmt.Errorf("update payment status: %w", err)
		}
		return nil
	}); err != nil {
		r.logger.Error("Failed to update payment status", zap.Error(err), zap.String("paymentID", paymentID), zap.Int32("statusID", statusID))
		return err
	}
	return nil
}

// UpdatePaymentID updates the payment ID
func (r *PaymentRepository) UpdatePaymentID(ctx context.Context, orderID uuid.UUID, paymentID string) error {
	if err := r.execTx(ctx, func(q *database.Queries) error {
		if err := q.UpdatePaymentID(ctx, database.UpdatePaymentIDParams{
			ID:        orderID,
			PaymentID: paymentID,
		}); err != nil {
			r.logger.Error("Failed to update payment ID", zap.Error(err), zap.String("orderID", orderID.String()), zap.String("paymentID", paymentID))
			return fmt.Errorf("update payment ID: %w", err)
		}
		return nil
	}); err != nil {
		r.logger.Error("Failed to update payment ID", zap.Error(err), zap.String("orderID", orderID.String()), zap.String("paymentID", paymentID))
		return err
	}
	return nil
}

// GetPaymentsByOrderID returns payments by order ID
func (r *PaymentRepository) GetPaymentsByOrderID(ctx context.Context, orderID uuid.UUID) ([]model.OrderPayment, error) {
	payments, err := r.Queries.GetPaymentsByOrderID(ctx, uuid.NullUUID{
		UUID:  orderID,
		Valid: true,
	})
	if err != nil {
		r.logger.Error("Failed to get payments by order ID", zap.Error(err), zap.String("orderID", orderID.String()))
		return nil, fmt.Errorf("get payments by order ID: %w", err)
	}

	var orderPayments []model.OrderPayment
	for _, payment := range payments {
		orderPayments = append(orderPayments, model.OrderPayment{
			ID:              payment.ID,
			OrderID:         payment.OrderID.UUID,
			PaymentID:       payment.PaymentID,
			Amount:          payment.Amount,
			CreatedAt:       payment.CreatedAt.Time,
			PaymentMethodID: payment.PaymentMethodID.Int32,
			PaymentStatusID: payment.PaymentStatusID.Int32,
		})
	}

	return orderPayments, nil
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
			ID:              payment.ID,
			OrderID:         payment.OrderID.UUID,
			PaymentID:       payment.PaymentID,
			Amount:          payment.Amount,
			CreatedAt:       payment.CreatedAt.Time,
			PaymentMethodID: payment.PaymentMethodID.Int32,
			PaymentStatusID: payment.PaymentStatusID.Int32,
		})

	}

	return orderPayments, nil
}
