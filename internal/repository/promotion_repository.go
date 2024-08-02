package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"strconv"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
)

type PromotionRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewPromotionRepository(db *sql.DB, logger *zap.Logger) *PromotionRepository {
	return &PromotionRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *PromotionRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
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

// CreatePromotion creates a new promotion
func (r *PromotionRepository) CreatePromotion(ctx context.Context, params *database.CreatePromotionParams) (*model.PromotionSchema, error) {
	var promotion *model.PromotionSchema
	if err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		data, err := q.CreatePromotion(ctx, *params)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				r.logger.Error("Failed to create promotion", zap.Error(err), zap.Any("promotionParams", params))
				return err
			default:
				r.logger.Error("Failed to create promotion", zap.Error(err), zap.Any("promotionParams", params))
				return fmt.Errorf("create promotion: %w", err)
			}
		}

		promotion = &model.PromotionSchema{
			ID:          data.ID,
			Title:       data.Title,
			Tagline:     data.Tagline.String,
			MainTitle:   data.MainTitle,
			SubTitle:    data.Subtitle,
			Description: data.Description.String,
			ImageUrl:    data.ImageUrl.String,
		}

		return nil
	}); err != nil {
		r.logger.Error("Create promotion transaction failed", zap.Error(err))
		return nil, err
	}
	return promotion, nil
}

// AddProductToPromotion adds a product to a promotion
func (r *PromotionRepository) AddProductToPromotion(ctx context.Context, promotionID, productID uuid.UUID) error {
	if err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.AddProductToPromotion(ctx, database.AddProductToPromotionParams{
			PromotionID: promotionID,
			ProductID:   productID,
		})
		if err != nil {
			r.logger.Error("add product to promotion failed", zap.Error(err))
			return fmt.Errorf("add product to promotion: %w", err)
		}
		return nil
	}); err != nil {
		r.logger.Error("Add product to promotion transaction failed", zap.Error(err))
		return err
	}
	return nil
}

// GetPromotionsWithProducts retrieves all promotions with their products
func (r *PromotionRepository) GetPromotionsWithProducts(ctx context.Context) ([]*model.Promotion, error) {
	promotions, err := r.Queries.GetPromotionsWithProducts(ctx)
	if err != nil {
		r.logger.Error("failed to get promotions", zap.Error(err))
		return nil, fmt.Errorf("failed to get promotions: %w", err)
	}

	var promotionList []*model.Promotion
	for _, promotion := range promotions {
		discount, err := strconv.ParseFloat(promotion.DiscountPercentage, 64)
		if err != nil {
			r.logger.Error("failed to parse discount percentage", zap.Error(err))
			return nil, fmt.Errorf("failed to parse discount percentage: %w", err)
		}

		promotionList = append(promotionList, &model.Promotion{
			Slug:              promotion.Slug,
			Tagline:           promotion.Tagline.String,
			MainTitle:         promotion.MainTitle,
			SubTitle:          promotion.Subtitle,
			Title:             promotion.Title,
			Description:       promotion.Description.String,
			Discount:          discount,
			PromotionImageUrl: promotion.PromotionImageUrl.String,
			ProductImageUrl:   promotion.ProductImageUrl.String,
		})
	}

	return promotionList, nil
}
