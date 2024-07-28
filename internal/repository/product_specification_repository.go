package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"weblineBackend/internal/database"
)

type ProductSpecificationRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

// NewProductSpecificationRepository initializes a new ProductSpecificationRepository with dependency injection for logging
func NewProductSpecificationRepository(db *sql.DB, logger *zap.Logger) *ProductSpecificationRepository {
	return &ProductSpecificationRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a transaction and rolls back if an error occurs
func (r *ProductSpecificationRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
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

// CreateProductSpecification creates a new product specification in the database and returns the created specification
func (r *ProductSpecificationRepository) CreateProductSpecification(
	ctx context.Context,
	params database.CreateProductSpecificationParams,
) (database.ProductSpecification, error) {
	var createdSpecification database.ProductSpecification
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		createdSpecification, err = q.CreateProductSpecification(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to create product specification: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to create product specification", zap.Error(err))
		return database.ProductSpecification{}, err
	}
	return createdSpecification, nil
}

// UpsertProductSpecification upserts a product specification in the database and returns the upserted specification
func (r *ProductSpecificationRepository) UpsertProductSpecification(
	ctx context.Context,
	params database.UpsertProductSpecificationParams,
) (database.ProductSpecification, error) {
	var upsertedSpec database.ProductSpecification
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		upsertedSpec, err = q.UpsertProductSpecification(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to upsert product specification: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to upsert product specification", zap.Error(err))
		return database.ProductSpecification{}, err
	}
	return upsertedSpec, nil
}

// GetProductSpecificationByID retrieves a product specification by its ID
func (r *ProductSpecificationRepository) GetProductSpecificationByID(
	ctx context.Context,
	id uuid.UUID,
) (database.ProductSpecification, error) {
	specification, err := r.Queries.GetProductSpecificationByID(ctx, id)
	if err != nil {
		r.logger.Error("failed to get product specification by ID", zap.Error(err))
		return database.ProductSpecification{}, fmt.Errorf("failed to get product specification by ID: %w", err)
	}
	return specification, nil
}

// ListProductSpecificationsByProductID retrieves product specifications by their product ID
func (r *ProductSpecificationRepository) ListProductSpecificationsByProductID(
	ctx context.Context,
	productID uuid.NullUUID,
) ([]database.ProductSpecification, error) {
	specifications, err := r.Queries.ListProductSpecificationsByProductID(ctx, productID)
	if err != nil {
		r.logger.Error("failed to list product specifications by product ID", zap.Error(err))
		return nil, fmt.Errorf("failed to list product specifications by product ID: %w", err)
	}
	return specifications, nil
}

// UpdateProductSpecification updates a product specification in the database and returns the updated specification
func (r *ProductSpecificationRepository) UpdateProductSpecification(
	ctx context.Context,
	params database.UpdateProductSpecificationParams,
) (database.ProductSpecification, error) {
	var updatedSpec database.ProductSpecification
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		updatedSpec, err = q.UpdateProductSpecification(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to update product specification: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to update product specification", zap.Error(err))
		return database.ProductSpecification{}, err
	}
	return updatedSpec, nil
}

// DeleteProductSpecification deletes a product specification from the database
func (r *ProductSpecificationRepository) DeleteProductSpecification(
	ctx context.Context,
	id uuid.UUID,
) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		if err := q.DeleteProductSpecification(ctx, id); err != nil {
			return fmt.Errorf("failed to delete product specification: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to delete product specification", zap.Error(err))
		return err
	}
	return nil
}
