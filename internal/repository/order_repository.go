package repository

import (
	"context"
	"database/sql"
	"fmt"
	"weblineBackend/internal/database"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrderRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewOrderRepository(db *sql.DB, logger *zap.Logger) *OrderRepository {
	return &OrderRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *OrderRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
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
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// CreateOrder creates a new order
func (r *OrderRepository) CreateOrder(ctx context.Context, order *database.CreateOrderParams) (*uuid.UUID, error) {
	var orderID uuid.UUID
	if err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		orderID, err = q.CreateOrder(ctx, *order)
		if err != nil {
			r.logger.Error("Failed to create order", zap.Error(err), zap.Any("orderParams", order))
			return fmt.Errorf("create order: %w", err)
		}
		return nil
	}); err != nil {
		r.logger.Error("Create order transaction failed", zap.Error(err))
		return nil, fmt.Errorf("create order transaction: %w", err)
	}
	r.logger.Info("Order created successfully", zap.String("orderID", orderID.String()))
	return &orderID, nil
}

// UpdateOrderStatus updates the status of an order
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error {
	if err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.UpdateOrderStatus(ctx, database.UpdateOrderStatusParams{
			ID:     orderID,
			Status: status,
		})
		if err != nil {
			r.logger.Error("update order status failed", zap.Error(err))
			return fmt.Errorf("update order status: %w", err)
		}
		return nil
	}); err != nil {
		r.logger.Error("update order status transaction failed", zap.Error(err))
		return fmt.Errorf("update order status transaction: %w", err)
	}
	return nil
}

// UpdateOrderPaymentStatus updates the payment status of an order
func (r *OrderRepository) UpdateOrderPaymentStatus(ctx context.Context, orderID uuid.UUID, paymentStatus string) error {
	if err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.UpdateOrderPaymentStatus(ctx, database.UpdateOrderPaymentStatusParams{
			ID:            orderID,
			PaymentStatus: paymentStatus,
		})
		if err != nil {
			r.logger.Error("update order payment status failed", zap.Error(err))
			return fmt.Errorf("update order payment status: %w", err)
		}
		return nil
	}); err != nil {
		r.logger.Error("update order payment status transaction failed", zap.Error(err))
		return fmt.Errorf("update order payment status transaction: %w", err)
	}
	return nil
}

// GetOrderById retrieves an order by its ID
func (r *OrderRepository) GetOrderById(ctx context.Context, orderID uuid.UUID) (*database.GetOrderByIdRow, error) {
	order, err := r.Queries.GetOrderById(ctx, orderID)
	if err != nil {
		r.logger.Error("get order by id failed", zap.Error(err))
		return nil, fmt.Errorf("get order by id: %w", err)
	}
	return &order, nil
}

// GetOrdersByUserId retrieves all orders for a specific user
func (r *OrderRepository) GetOrdersByUserId(ctx context.Context, userID uuid.NullUUID) ([]database.GetOrdersByUserIdRow, error) {
	orders, err := r.Queries.GetOrdersByUserId(ctx, userID)
	if err != nil {
		r.logger.Error("get orders by user id failed", zap.Error(err))
		return nil, fmt.Errorf("get orders by user id: %w", err)
	}
	return orders, nil
}

// GetOrdersByGuestCheckoutId retrieves all orders for a specific guest checkout
func (r *OrderRepository) GetOrdersByGuestCheckoutId(ctx context.Context, guestCheckoutID uuid.NullUUID) ([]database.GetOrdersByGuestCheckoutIdRow, error) {
	orders, err := r.Queries.GetOrdersByGuestCheckoutId(ctx, guestCheckoutID)
	if err != nil {
		r.logger.Error("get orders by guest checkout id failed", zap.Error(err))
		return nil, fmt.Errorf("get orders by guest checkout id: %w", err)
	}
	return orders, nil
}

// CreateGuestCheckout creates a guest checkout
func (r *OrderRepository) CreateGuestCheckout(ctx context.Context, guestCheckout *database.CreateGuestCheckoutParams) (*uuid.UUID, error) {
	var orderID uuid.UUID
	if err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		orderID, err = q.CreateGuestCheckout(ctx, *guestCheckout)
		if err != nil {
			r.logger.Error("create guest checkout failed", zap.Error(err))
			return fmt.Errorf("create guest checkout: %w", err)
		}
		return nil
	}); err != nil {
		r.logger.Error("create guest checkout transaction failed", zap.Error(err))
		return nil, fmt.Errorf("create guest checkout transaction: %w", err)
	}
	return &orderID, nil
}

// GetOrderIDsByUserID retrieves all order IDs for a specific user
func (r *OrderRepository) GetOrderIDsByUserID(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	orderIDs, err := r.Queries.GetOrderIDsByUserID(ctx, uuid.NullUUID{UUID: userID, Valid: true})
	if err != nil {
		r.logger.Error("get order IDs by user ID failed", zap.Error(err))
		return nil, fmt.Errorf("get order IDs by user ID: %w", err)
	}

	r.logger.Info("Order IDs retrieved successfully", zap.Any("orderIDs", orderIDs))
	return orderIDs, nil
}

// GetUserOrGuestCheckoutNameByOrderID retrieves the name of the user or guest checkout for a specific order
func (r *OrderRepository) GetUserOrGuestCheckoutNameByOrderID(ctx context.Context, orderID uuid.UUID) (*database.GetUserOrGuestCheckoutNameByOrderIDRow, error) {
	names, err := r.Queries.GetUserOrGuestCheckoutNameByOrderID(ctx, orderID)
	if err != nil {
		r.logger.Error("get user or guest checkout name by order ID failed", zap.Error(err))
		return nil, fmt.Errorf("get user or guest checkout name by order ID: %w", err)
	}

	r.logger.Info("User or guest checkout name retrieved successfully", zap.Any("names", names))
	return &names, nil
}
