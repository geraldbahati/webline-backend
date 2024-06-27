package repository

import (
	"context"
	"database/sql"
	"fmt"
	"weblineBackend/internal/database"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProductSizeRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

// NewProductSizeRepository initializes a new ProductSizeRepository with dependency injection for logging
func NewProductSizeRepository(db *sql.DB, logger *zap.Logger) *ProductSizeRepository {
	return &ProductSizeRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx is a helper function to execute a transaction
func (r *ProductSizeRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
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

// CreateProductSize stores a product size in the database and returns the created product size
func (r *ProductSizeRepository) CreateProductSize(
	ctx context.Context,
	params database.CreateProductSizeParams,
) (*database.CreateProductSizeRow, error) {
	var productSize database.CreateProductSizeRow
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		productSize, err = q.CreateProductSize(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to create product size: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to create product size", zap.Error(err))
		return nil, err
	}
	return &productSize, nil
}

// GetProductSizeByID retrieves a product size by its ID
func (r *ProductSizeRepository) GetProductSizeByID(
	ctx context.Context,
	id uuid.UUID,
) (*database.GetProductSizeByIDRow, error) {
	productSize, err := r.Queries.GetProductSizeByID(ctx, id)
	if err != nil {
		r.logger.Error("failed to get product size", zap.Error(err))
		return nil, fmt.Errorf("failed to get product size: %w", err)
	}

	r.logger.Info("product size retrieved", zap.String("id", productSize.ID.String()))
	return &productSize, nil
}

// GetProductSizesByProductID retrieves all product sizes for a given product ID
func (r *ProductSizeRepository) GetProductSizesByProductID(
	ctx context.Context,
	productID uuid.NullUUID,
) (*[]database.ListProductSizesByProductIDRow, error) {
	productSizes, err := r.Queries.ListProductSizesByProductID(ctx, productID)
	if err != nil {
		r.logger.Error("failed to get product sizes", zap.Error(err))
		return nil, err
	}
	return &productSizes, nil
}

// UpdateProductSize updates a product size in the database
func (r *ProductSizeRepository) UpdateProductSize(
	ctx context.Context,
	params database.UpdateProductSizeParams,
) (*database.UpdateProductSizeRow, error) {
	var productSize database.UpdateProductSizeRow
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		productSize, err = q.UpdateProductSize(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to update product size: %w", err)
		}

		r.logger.Info("product size updated", zap.String("id", productSize.ID.String()))
		return nil
	})
	if err != nil {
		r.logger.Error("failed to update product size", zap.Error(err))
		return nil, fmt.Errorf("failed to update product size: %w", err)
	}

	r.logger.Info("product size updated", zap.String("id", productSize.ID.String()))
	return &productSize, nil
}

// DeleteProductSize deletes a product size from the database
func (r *ProductSizeRepository) DeleteProductSize(
	ctx context.Context,
	id uuid.UUID,
) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		if err := q.DeleteProductSize(ctx, id); err != nil {
			return fmt.Errorf("failed to delete product size: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to delete product size", zap.Error(err))
		return err
	}
	return nil
}

// GetAvailableColorsByCategoryID retrieves all available colors for a given category ID
func (r *ProductSizeRepository) GetAvailableColorsByCategoryID(
	ctx context.Context,
	categoryID uuid.UUID,
) ([]string, error) {
	colors, err := r.Queries.GetAvailableSizesByParentCategoryID(ctx, categoryID)
	if err != nil {
		r.logger.Error("failed to get available colors", zap.Error(err))
		return nil, err
	}
	return colors, nil
}

// GetAllProductSizes retrieves all product sizes
func (r *ProductSizeRepository) GetAllProductSizes(
	ctx context.Context,
) ([]database.GetAllSizesRow, error) {
	productSizes, err := r.Queries.GetAllSizes(ctx)
	if err != nil {
		r.logger.Error("failed to get product sizes", zap.Error(err))
		return nil, err
	}
	return productSizes, nil
}
