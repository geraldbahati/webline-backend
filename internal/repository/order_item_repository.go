package repository

import (
	"context"
	"database/sql"
	"fmt"

	"weblineBackend/internal/database"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// OrderItemRepository handles CRUD operations for order items.
type OrderItemRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

// NewOrderItemRepository initializes a new OrderItemRepository.
func NewOrderItemRepository(db *sql.DB, logger *zap.Logger) *OrderItemRepository {
	return &OrderItemRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function.
// It ensures proper commit or rollback based on the function's outcome.
func (r *OrderItemRepository) execTx(ctx context.Context, fn func(*database.Queries) error) (err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return fmt.Errorf("begin transaction: %w", err)
	}

	q := database.New(tx)
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			r.logger.Panic("transaction panicked, rolling back", zap.Any("panic", p))
			panic(p) // Re-throw panic after rollback
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				r.logger.Error("rollback failed", zap.Error(rbErr))
				err = fmt.Errorf("transaction rollback failed: %v, original error: %w", rbErr, err)
			} else {
				r.logger.Warn("transaction rolled back due to error", zap.Error(err))
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				r.logger.Error("commit failed", zap.Error(commitErr))
				err = fmt.Errorf("commit transaction: %w", commitErr)
			}
		}
	}()

	err = fn(q)
	return err
}

// CreateOrderItem creates a new order item within a transaction.
// Returns the UUID of the created order item.
func (r *OrderItemRepository) CreateOrderItem(ctx context.Context, item *database.CreateOrderItemParams) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		id, err = q.CreateOrderItem(ctx, *item)
		if err != nil {
			r.logger.Error("failed to create order item", zap.Error(err), zap.Any("item", item))
			return fmt.Errorf("create order item: %w", err)
		}
		return nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("CreateOrderItem transaction failed: %w", err)
	}
	return id, nil
}

// GetOrderItemsByOrderId retrieves order items by the given order ID.
func (r *OrderItemRepository) GetOrderItemsByOrderId(ctx context.Context, orderID uuid.UUID) ([]database.GetOrderItemsByOrderIdRow, error) {
	items, err := r.Queries.GetOrderItemsByOrderId(ctx, uuid.NullUUID{
		UUID:  orderID,
		Valid: true,
	})
	if err != nil {
		r.logger.Error("failed to get order items by order ID", zap.Error(err), zap.String("orderID", orderID.String()))
		return nil, fmt.Errorf("get order items by order ID: %w", err)
	}
	return items, nil
}

// CreateOrderItemOption creates a new option for an order item within a transaction.
func (r *OrderItemRepository) CreateOrderItemOption(ctx context.Context, option *database.CreateOrderItemOptionParams) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		if err := q.CreateOrderItemOption(ctx, *option); err != nil {
			r.logger.Error("failed to create order item option", zap.Error(err), zap.Any("option", option))
			return fmt.Errorf("create order item option: %w", err)
		}

		r.logger.Info("successfully created order item option", zap.Any("option", option))
		return nil
	})
	if err != nil {
		return fmt.Errorf("CreateOrderItemOption transaction failed: %w", err)
	}
	return nil
}


// CreateOrderItems creates multiple order items within a transaction.
func (r *OrderItemRepository) CreateOrderItems(ctx context.Context, items database.CreateOrderItemsParams) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		if err := q.CreateOrderItems(ctx, items); err != nil {
			r.logger.Error("failed to create order items", zap.Error(err), zap.Any("items", items))
			return fmt.Errorf("create order items: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("CreateOrderItems transaction failed: %w", err)
	}
	return nil
}