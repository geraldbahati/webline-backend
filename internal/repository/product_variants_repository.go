package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"weblineBackend/internal/database"
)

type ProductVariantRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

// NewProductVariantRepository initializes a new ProductVariantRepository with dependency injection for logging
func NewProductVariantRepository(db *sql.DB, logger *zap.Logger) *ProductVariantRepository {
	return &ProductVariantRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *ProductVariantRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
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

// CreateProductVariant stores a product variant in the database and returns the created variant
func (r *ProductVariantRepository) CreateProductVariant(
	ctx context.Context,
	variant database.CreateProductVariantParams,
) (database.ProductVariant, error) {
	var createdVariant database.ProductVariant
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		createdVariant, err = q.CreateProductVariant(ctx, variant)
		if err != nil {
			return fmt.Errorf("failed to create product variant: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to create product variant", zap.Error(err))
		return database.ProductVariant{}, err
	}
	return createdVariant, nil
}

// GetProductVariantByID retrieves a product variant by its ID
func (r *ProductVariantRepository) GetProductVariantByID(
	ctx context.Context,
	id uuid.UUID,
) (database.ProductVariant, error) {
	variant, err := r.Queries.GetProductVariantByID(ctx, id)
	if err != nil {
		r.logger.Error("failed to get product variant by ID", zap.Error(err))
		return database.ProductVariant{}, fmt.Errorf("failed to get product variant by ID: %w", err)
	}
	return variant, nil
}

// ListProductVariantsByProductID retrieves all product variants by product ID
func (r *ProductVariantRepository) ListProductVariantsByProductID(
	ctx context.Context,
	productID uuid.NullUUID,
) ([]database.ProductVariant, error) {
	variants, err := r.Queries.ListProductVariantsByProductID(ctx, productID)
	if err != nil {
		r.logger.Error("failed to list product variants by product ID", zap.Error(err))
		return nil, fmt.Errorf("failed to list product variants by product ID: %w", err)
	}
	return variants, nil
}

// UpdateProductVariant updates a product variant in the database and returns the updated variant
func (r *ProductVariantRepository) UpdateProductVariant(
	ctx context.Context,
	params database.UpdateProductVariantParams,
) (database.ProductVariant, error) {
	var updatedVariant database.ProductVariant
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		updatedVariant, err = q.UpdateProductVariant(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to update product variant: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to update product variant", zap.Error(err))
		return database.ProductVariant{}, err
	}
	return updatedVariant, nil
}

// DeleteProductVariant deletes a product variant from the database
func (r *ProductVariantRepository) DeleteProductVariant(
	ctx context.Context,
	id uuid.UUID,
) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		if err := q.DeleteProductVariant(ctx, id); err != nil {
			return fmt.Errorf("failed to delete product variant: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to delete product variant", zap.Error(err))
		return err
	}
	return nil
}
