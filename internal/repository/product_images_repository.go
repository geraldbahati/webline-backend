package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"weblineBackend/internal/database"
)

type ProductImageRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

// NewProductImageRepository initializes a new ProductImageRepository with dependency injection for logging
func NewProductImageRepository(db *sql.DB, logger *zap.Logger) *ProductImageRepository {
	return &ProductImageRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx is a helper function to execute a transaction
func (r *ProductImageRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	q := database.New(tx)
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			r.logger.Error("transaction failed, rolling back", zap.Error(err))
			if rbErr := tx.Rollback(); rbErr != nil {
				r.logger.Error("rollback failed", zap.Error(rbErr))
				err = fmt.Errorf("rollback transaction: %w", rbErr)
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("commit transaction: %w", commitErr)
			}
		}
	}()

	err = fn(q)
	return err
}

// CreateProductImage stores a product image in the database and returns the created product image
func (r *ProductImageRepository) CreateProductImage(
	ctx context.Context,
	params database.CreateProductImageParams,
) (database.ProductImage, error) {
	var productImage database.ProductImage
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		productImage, err = q.CreateProductImage(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to create product image: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to create product image", zap.Error(err))
		return database.ProductImage{}, err
	}
	return productImage, nil
}

// GetProductImageByID retrieves a product image by its ID
func (r *ProductImageRepository) GetProductImageByID(
	ctx context.Context,
	id uuid.UUID,
) (database.ProductImage, error) {
	productImage, err := r.Queries.GetProductImageByID(ctx, id)
	if err != nil {
		r.logger.Error("failed to get product image by ID", zap.Error(err))
		return database.ProductImage{}, fmt.Errorf("failed to get product image by ID: %w", err)
	}
	return productImage, nil
}

// ListProductImagesByProductID retrieves a list of product images by product ID
func (r *ProductImageRepository) ListProductImagesByProductID(
	ctx context.Context,
	productID uuid.NullUUID,
) ([]database.ProductImage, error) {
	productImages, err := r.Queries.ListProductImagesByProductID(ctx, productID)
	if err != nil {
		r.logger.Error("failed to list product images by product ID", zap.Error(err))
		return nil, fmt.Errorf("failed to list product images by product ID: %w", err)
	}
	return productImages, nil
}

// UpdateProductImage updates a product image in the database and returns the updated image
func (r *ProductImageRepository) UpdateProductImage(
	ctx context.Context,
	params database.UpdateProductImageParams,
) (database.ProductImage, error) {
	var updatedImage database.ProductImage
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		updatedImage, err = q.UpdateProductImage(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to update product image: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to update product image", zap.Error(err))
		return database.ProductImage{}, err
	}
	return updatedImage, nil
}

// DeleteProductImage deletes a product image from the database
func (r *ProductImageRepository) DeleteProductImage(
	ctx context.Context,
	id uuid.UUID,
) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		if err := q.DeleteProductImage(ctx, id); err != nil {
			return fmt.Errorf("failed to delete product image: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to delete product image", zap.Error(err))
		return err
	}
	return nil
}

// GetImageKeysByProductID retrieves a list of image keys by product ID
func (r *ProductImageRepository) GetImageKeysByProductID(
	ctx context.Context,
	productID uuid.UUID,
) ([]string, error) {
	imageKeys, err := r.Queries.GetImageKeysByProductID(ctx, uuid.NullUUID{UUID: productID, Valid: true})
	if err != nil {
		r.logger.Error("failed to get image keys by product ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get image keys by product ID: %w", err)
	}
	return imageKeys, nil
}
