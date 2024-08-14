package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
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
func (r *PromotionRepository) execTx(ctx context.Context, fn func(*database.Queries) error) (err error) {
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

// GetPromotions retrieves all promotions
func (r *PromotionRepository) GetPromotions(ctx context.Context) ([]*model.Promotion, error) {
	promotions, err := r.Queries.GetPromotions(ctx)
	if err != nil {
		r.logger.Error("failed to get promotions", zap.Error(err))
		return nil, fmt.Errorf("failed to get promotions: %w", err)
	}

	var promotionList []*model.Promotion
	for _, promotion := range promotions {

		promotionList = append(promotionList, &model.Promotion{
			Name:        promotion.Name,
			ProductSlug: promotion.Productslug,
			Description: promotion.Description.String,
			Slug:        promotion.Slug,
			ImageUrl:    promotion.Imageurl.String,
		})
	}

	r.logger.Info("Promotions retrieved successfully")
	return promotionList, nil
}

// GetV2Promotions retrieves all promotions for dashboard
func (r *PromotionRepository) GetV2Promotions(ctx context.Context) ([]*model.V2Promotion, error) {
	promotions, err := r.Queries.GetV2Promotions(ctx)
	if err != nil {
		r.logger.Error("failed to get promotions", zap.Error(err))
		return nil, fmt.Errorf("failed to get promotions: %w", err)
	}

	var promotionList []*model.V2Promotion
	for _, promotion := range promotions {

		promotionList = append(promotionList, &model.V2Promotion{
			Slug:             promotion.Slug,
			ID:               promotion.ID,
			Name:             promotion.Name,
			Type:             promotion.Type,
			ImageUrl:         promotion.ImageUrl.String,
			NumberOfProducts: promotion.Numberofproducts,
			Status:           promotion.Status,
			StartDate:        promotion.Startdate,
			EndDate:          promotion.Enddate,
		})
	}

	r.logger.Info("Promotions retrieved successfully")
	return promotionList, nil
}

// GetPromotionBySlug retrieves a promotion by its slug
func (r *PromotionRepository) GetPromotionBySlug(ctx context.Context, slug string) (*model.PromotionSchema, error) {
	promotion, err := r.Queries.GetPromotionBySlug(ctx, slug)
	if err != nil {
		r.logger.Error("failed to get promotion", zap.Error(err))
		return nil, fmt.Errorf("failed to get promotion: %w", err)
	}

	return &model.PromotionSchema{
		ID:          promotion.ID,
		Title:       promotion.Title,
		Description: promotion.Description.String,
		ImageUrl:    promotion.ImageUrl.String,
		StartDate:   promotion.StartDate,
		EndDate:     promotion.EndDate,
		Slug:        promotion.Slug,
		Status:      promotion.Status,
	}, nil
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

// UpdatePromotionImage updates the promotion image
func (r *PromotionRepository) UpdatePromotionImage(ctx context.Context, promotionID uuid.UUID, imageUrl string) error {
	if imageUrl == "" {
		return nil
	}

	if err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.UpdatePromotionImage(ctx, database.UpdatePromotionImageParams{
			ID:       promotionID,
			ImageUrl: sql.NullString{String: imageUrl, Valid: imageUrl != ""},
		})
		if err != nil {
			r.logger.Error("update promotion image failed", zap.Error(err))
			return fmt.Errorf("update promotion image: %w", err)
		}
		return nil
	}); err != nil {
		r.logger.Error("Update promotion image transaction failed", zap.Error(err))
		return err
	}
	return nil
}

// UpdatePromotion updates a promotion
func (r *PromotionRepository) UpdatePromotion(ctx context.Context, params *database.UpdatePromotionParams) error {
	if err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.UpdatePromotion(ctx, *params)
		if err != nil {
			r.logger.Error("update promotion failed", zap.Error(err))
			return fmt.Errorf("update promotion: %w", err)
		}
		return nil
	}); err != nil {
		r.logger.Error("Update promotion transaction failed", zap.Error(err))
		return err
	}
	return nil
}

// AddProductsToPromotion adds products to a promotion
func (r *PromotionRepository) AddProductsToPromotion(ctx context.Context, promotionID uuid.UUID, productIDs []uuid.UUID) error {
	if err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.AddProductsToPromotion(ctx, database.AddProductsToPromotionParams{
			PromotionID: promotionID,
			Column2:     productIDs,
		})
		if err != nil {
			r.logger.Error("failed to add products to promotion", zap.String("promotionID", promotionID.String()), zap.Int("productCount", len(productIDs)), zap.Error(err))
			return fmt.Errorf("add products to promotion: %w", err)
		}
		return nil
	}); err != nil {
		r.logger.Error("add products to promotion transaction failed", zap.String("promotionID", promotionID.String()), zap.Error(err))
		return err
	}

	r.logger.Info("Products added to promotion successfully", zap.String("promotionID", promotionID.String()), zap.Int("productCount", len(productIDs)))
	return nil
}

// GetProductIDsByPromotionID retrieves product IDs by promotion ID
func (r *PromotionRepository) GetProductIDsByPromotionID(ctx context.Context, promotionID uuid.UUID) ([]uuid.UUID, error) {
	productIDs, err := r.Queries.GetProductIDsByPromotionID(ctx, promotionID)
	if err != nil {
		r.logger.Error("failed to get product IDs by promotion ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get product IDs by promotion ID: %w", err)
	}

	return productIDs, nil
}

// RemoveProductsFromPromotion removes products from a promotion
func (r *PromotionRepository) RemoveProductsFromPromotion(ctx context.Context, promotionID uuid.UUID, productIDs []uuid.UUID) error {
	if err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.RemoveProductsFromPromotion(ctx, database.RemoveProductsFromPromotionParams{
			PromotionID: promotionID,
			Column2:     productIDs,
		})
		if err != nil {
			r.logger.Error("failed to remove products from promotion", zap.String("promotionID", promotionID.String()), zap.Int("productCount", len(productIDs)), zap.Error(err))
			return fmt.Errorf("remove products from promotion: %w", err)
		}
		return nil
	}); err != nil {
		r.logger.Error("remove products from promotion transaction failed", zap.String("promotionID", promotionID.String()), zap.Error(err))
		return err
	}

	r.logger.Info("Products removed from promotion successfully", zap.String("promotionID", promotionID.String()), zap.Int("productCount", len(productIDs)))
	return nil
}
