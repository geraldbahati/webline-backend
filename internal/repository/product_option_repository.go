package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"weblineBackend/internal/database"
)

type ProductOptionRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewProductOptionRepository(db *sql.DB, logger *zap.Logger) *ProductOptionRepository {
	return &ProductOptionRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *ProductOptionRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
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

// CreateProductOption stores a product option in the database and returns the created product option
func (r *ProductOptionRepository) CreateProductOption(
	ctx context.Context,
	params database.CreateProductOptionParams,
) (database.ProductOption, error) {
	var productOption database.ProductOption
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		productOption, err = q.CreateProductOption(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to create product option: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to create product option", zap.Error(err))
		return database.ProductOption{}, err
	}
	return productOption, nil
}

// GetProductOptionsByProductID retrieves all product options for a given product ID
func (r *ProductOptionRepository) GetProductOptionsByProductID(
	ctx context.Context,
	productID uuid.NullUUID,
) ([]database.ProductOption, error) {
	productOptions, err := r.ListProductOptionsByProductID(ctx, productID)
	if err != nil {
		r.logger.Error("failed to get product options", zap.Error(err))
		return nil, err
	}
	return productOptions, nil
}

// GetProductOptionByID retrieves a product option by its ID
func (r *ProductOptionRepository) GetProductOptionByID(
	ctx context.Context,
	id uuid.UUID,
) (database.ProductOption, error) {
	productOption, err := r.Queries.GetProductOptionByID(ctx, id)
	if err != nil {
		r.logger.Error("failed to get product option by ID", zap.Error(err))
		return database.ProductOption{}, fmt.Errorf("failed to get product option by ID: %w", err)
	}
	return productOption, nil
}

// UpdateProductOption updates a product option in the database
func (r *ProductOptionRepository) UpdateProductOption(
	ctx context.Context,
	params database.UpdateProductOptionParams,
) (database.ProductOption, error) {
	var productOption database.ProductOption
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		productOption, err = q.UpdateProductOption(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to create product option: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to create product option", zap.Error(err))
		return database.ProductOption{}, err
	}
	return productOption, nil
}

// DeleteProductOption deletes a product option from the database
func (r *ProductOptionRepository) DeleteProductOption(
	ctx context.Context,
	id uuid.UUID,
) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		if err := q.DeleteProductOption(ctx, id); err != nil {
			return fmt.Errorf("failed to delete product option: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to delete product option", zap.Error(err))
		return fmt.Errorf("failed to delete product option: %w", err)
	}
	return nil
}
