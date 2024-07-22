package repository

import (
	"context"
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"strconv"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
)

type ProductAnalyticRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewProductAnalyticRepository(db *sql.DB, logger *zap.Logger) *ProductAnalyticRepository {
	return &ProductAnalyticRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *ProductAnalyticRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
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

// GetBestSellerProducts returns the products
func (r *ProductAnalyticRepository) GetBestSellerProducts(ctx context.Context, limit int32) ([]*model.Product, error) {
	rows, err := r.Queries.GetBestSellerProducts(ctx, limit)
	if err != nil {
		r.logger.Error("Failed to get best seller products", zap.Error(err))
		return nil, fmt.Errorf("get best seller products: %w", err)
	}

	var products []*model.Product
	for _, row := range rows {
		products = append(products, &model.Product{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description.String,
			Price:       row.Price,
			Stock:       row.Stock.Int32,
			CategoryID:  row.CategoryID,
			IsActive:    row.IsActive.Bool,
			Featured:    row.Featured.Bool,
			Slug:        row.Slug.String,
		})
	}

	r.logger.Info("Best seller products retrieved successfully")
	return products, nil
}

// GetFeaturedProducts returns the featured products
func (r *ProductAnalyticRepository) GetFeaturedProducts(ctx context.Context, limit int32) ([]*model.Product, error) {
	rows, err := r.Queries.GetFeaturedProducts(ctx, limit)
	if err != nil {
		r.logger.Error("Failed to get featured products", zap.Error(err))
		return nil, fmt.Errorf("get featured products: %w", err)
	}

	var products []*model.Product
	for _, row := range rows {
		products = append(products, &model.Product{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description.String,
			Price:       row.Price,
			Stock:       row.Stock.Int32,
			CategoryID:  row.CategoryID,
			IsActive:    row.IsActive.Bool,
			Featured:    row.Featured.Bool,
			Slug:        row.Slug.String,
		})
	}

	r.logger.Info("Featured products retrieved successfully")
	return products, nil
}

// GetNewArrivalProducts returns the new arrival products
func (r *ProductAnalyticRepository) GetNewArrivalProducts(ctx context.Context, limit int32) ([]*model.Product, error) {
	rows, err := r.Queries.GetNewArrivalProducts(ctx, limit)
	if err != nil {
		r.logger.Error("Failed to get new arrival products", zap.Error(err))
		return nil, fmt.Errorf("get new arrival products: %w", err)
	}

	var products []*model.Product
	for _, row := range rows {
		products = append(products, &model.Product{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description.String,
			Price:       row.Price,
			Stock:       row.Stock.Int32,
			CategoryID:  row.CategoryID, Slug: row.Slug.String,
			IsActive: row.IsActive.Bool,
			Featured: row.Featured.Bool,
		})
	}

	r.logger.Info("New arrival products retrieved successfully")
	return products, nil
}

// GetDailyDealsProducts returns the daily deals products
func (r *ProductAnalyticRepository) GetDailyDealsProducts(ctx context.Context) ([]*model.Product, error) {
	rows, err := r.Queries.GetDailyDeals(ctx)
	if err != nil {
		r.logger.Error("Failed to get daily deals products", zap.Error(err))
		return nil, fmt.Errorf("get daily deals products: %w", err)
	}

	var products []*model.Product
	for _, row := range rows {
		discount, err := strconv.ParseFloat(row.DiscountPercentage, 64)
		if err != nil {
			r.logger.Error("Failed to parse discount percent", zap.Error(err))
			return nil, fmt.Errorf("parse discount percent: %w", err)
		}

		imageUrl := ""
		if row.ImageUrl.Valid {
			imageUrl = row.ImageUrl.String
		}

		products = append(products, &model.Product{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description.String,
			Price:       row.Price,
			Stock:       row.Stock.Int32,
			CategoryID:  row.CategoryID, Slug: row.Slug.String,
			IsActive:        row.IsActive.Bool,
			Featured:        row.Featured.Bool,
			ImageURL:        imageUrl,
			DiscountPercent: discount,
		})
	}

	r.logger.Info("Daily deals products retrieved successfully")
	return products, nil
}
