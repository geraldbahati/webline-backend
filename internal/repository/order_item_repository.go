package repository

import (
	"context"
	"database/sql"
	"fmt"
	"weblineBackend/internal/database"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrderItemRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewOrderItemRepository(db *sql.DB, logger *zap.Logger) *OrderItemRepository {
	return &OrderItemRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *OrderItemRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
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

// CreateOrderItem creates a new order item
func (r *OrderItemRepository) CreateOrderItem(ctx context.Context, item *database.CreateOrderItemParams) (*uuid.UUID, error) {
	var id uuid.UUID
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		id, err = q.CreateOrderItem(ctx, *item)
		if err != nil {
			r.logger.Error("create order item failed", zap.Error(err))
			return fmt.Errorf("create order item: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("create order item: %w", err)
	}
	return &id, nil
}

// GetOrderItemsByOrderId retrieves order items by order ID
func (r *OrderItemRepository) GetOrderItemsByOrderId(ctx context.Context, orderID uuid.NullUUID) ([]database.OrderItem, error) {
	items, err := r.Queries.GetOrderItemsByOrderId(ctx, orderID)
	if err != nil {
		r.logger.Error("get order items by order ID failed", zap.Error(err))
		return nil, fmt.Errorf("get order items by order ID: %w", err)
	}
	return items, nil
}

// CreateOrderItemOption creates a new order item option
func (r *OrderItemRepository) CreateOrderItemOption(ctx context.Context, option *database.CreateOrderItemOptionParams) error {
	return r.execTx(ctx, func(q *database.Queries) error {
		if err := q.CreateOrderItemOption(ctx, *option); err != nil {
			r.logger.Error("create order item option failed", zap.Error(err))
			return fmt.Errorf("create order item option: %w", err)
		}

		r.logger.Info("create order item option succeeded", zap.Any("option", option))
		return nil
	})
}
