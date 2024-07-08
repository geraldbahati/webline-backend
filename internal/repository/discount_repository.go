package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"weblineBackend/internal/database"
)

type DiscountRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewDiscountRepository(db *sql.DB, logger *zap.Logger) *DiscountRepository {
	return &DiscountRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *DiscountRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
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

// CreateDiscount creates a new discount
func (r *DiscountRepository) CreateDiscount(ctx context.Context, discount *database.CreateDiscountParams) (*database.Discount, error) {
	var createdDiscount database.Discount

	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		createdDiscount, err = q.CreateDiscount(ctx, *discount)
		if err != nil {
			return fmt.Errorf("create discount: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to create discount", zap.Error(err))
		return nil, err
	}

	return &createdDiscount, nil
}

// GetDiscountByProductID returns a discount by product id
func (r *DiscountRepository) GetDiscountByProductID(ctx context.Context, productID *uuid.UUID) (*database.Discount, error) {
	discount, err := r.Queries.GetDiscountByProductID(ctx, uuid.NullUUID{UUID: *productID, Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No discount found, return nil without error
			return nil, nil
		}
		r.logger.Error("failed to get discount by product id", zap.Error(err))
		return nil, err
	}

	return &discount, nil
}
