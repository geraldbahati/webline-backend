package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"weblineBackend/internal/database"
)

type ProductColourRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

// NewProductColourRepository initializes a new ProductColourRepository with dependency injection for logging
func NewProductColourRepository(db *sql.DB, logger *zap.Logger) *ProductColourRepository {
	return &ProductColourRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx is a helper function to execute a transaction
func (r *ProductColourRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
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

// CreateProductColor stores a product colour in the database and returns the created product colour
func (r *ProductColourRepository) CreateProductColor(
	ctx context.Context,
	params database.CreateProductColorParams,
) (database.ProductColor, error) {
	var productColour database.ProductColor
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		productColour, err = q.CreateProductColor(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to create product colour: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to create product colour", zap.Error(err))
		return database.ProductColor{}, err
	}
	return productColour, nil
}

// GetProductColorByID retrieves a product colour by its ID
func (r *ProductColourRepository) GetProductColorByID(
	ctx context.Context,
	id int,
) (database.ProductColor, error) {
	productColour, err := r.GetProductColorByID(ctx, id)
	if err != nil {
		r.logger.Error("failed to get product colour", zap.Error(err))
		return database.ProductColor{}, err
	}
	return productColour, nil
}

// GetProductColorsByProductID retrieves all product colours for a given product ID
func (r *ProductColourRepository) GetProductColorsByProductID(
	ctx context.Context,
	productID uuid.NullUUID,
) ([]database.ProductColor, error) {
	productColours, err := r.ListProductColorsByProductID(ctx, productID)
	if err != nil {
		r.logger.Error("failed to get product colours", zap.Error(err))
		return nil, err
	}
	return productColours, nil
}

// UpdateProductColor updates a product colour in the database and returns the updated product colour
func (r *ProductColourRepository) UpdateProductColor(
	ctx context.Context,
	params database.UpdateProductColorParams,
) (database.ProductColor, error) {
	var productColour database.ProductColor
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		productColour, err = q.UpdateProductColor(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to update product colour: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to update product colour", zap.Error(err))
		return database.ProductColor{}, err
	}
	return productColour, nil
}

// DeleteProductColor deletes a product colour from the database
func (r *ProductColourRepository) DeleteProductColor(
	ctx context.Context,
	id uuid.UUID,
) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		if err := q.DeleteProductColor(ctx, id); err != nil {
			return fmt.Errorf("failed to delete product colour: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to delete product colour", zap.Error(err))
		return fmt.Errorf("failed to delete product colour: %w", err)
	}
	return nil
}
