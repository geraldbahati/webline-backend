package sqlc

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"weblineBackend/internal/database"
)

type ProductAttributeRepoImpl struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewProductAttributeRepoImpl(db *sql.DB, logger *zap.Logger) *ProductAttributeRepoImpl {
	return &ProductAttributeRepoImpl{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *ProductAttributeRepoImpl) execTx(ctx context.Context, fn func(*database.Queries) error) (err error) {
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

// CreateProductAttribute creates a new product attribute
func (r *ProductAttributeRepoImpl) CreateProductAttribute(ctx context.Context, name, attributeType string) (*uuid.UUID, error) {
	var id uuid.UUID
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		id, err = q.CreateProductAttribute(ctx, database.CreateProductAttributeParams{
			Name:          name,
			AttributeType: database.AttributeTypeEnum(attributeType),
		})

		r.logger.Debug("created product attribute", zap.String("name", name), zap.String("attributeType", attributeType))
		return err
	})
	if err != nil {
		r.logger.Error("failed to create product attribute", zap.Error(err))
		return nil, err
	}

	r.logger.Info("product attribute created", zap.String("name", name), zap.String("attributeType", attributeType))
	return &id, err
}

// CreateProductAttributeValue creates a new product attribute value
func (r *ProductAttributeRepoImpl) CreateProductAttributeValue(ctx context.Context, attributeID uuid.UUID, categoryID uuid.NullUUID, value, hexValue string) (*uuid.UUID, error) {
	var id uuid.UUID
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		id, err = q.CreateProductAttributeValue(ctx, database.CreateProductAttributeValueParams{
			AttributeID: uuid.NullUUID{
				UUID:  attributeID,
				Valid: true,
			},
			CategoryID: categoryID,
			Value:      value,
			HexValue: sql.NullString{
				String: hexValue,
				Valid:  hexValue != "",
			},
		})

		r.logger.Debug("created product attribute value", zap.String("value", value), zap.String("hexaValue", hexaValue))
		return err
	})
	if err != nil {
		r.logger.Error("failed to create product attribute value", zap.Error(err))
		return nil, err
	}

	r.logger.Info("product attribute value created", zap.String("value", value), zap.String("hexaValue", hexaValue))
	return &id, err
}

// CreateProductToAttributeValue creates a new product to attribute value
func (r *ProductAttributeRepoImpl) CreateProductToAttributeValue(ctx context.Context, productID uuid.UUID, attributeValueID uuid.UUID) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.CreateProductToAttributeValue(ctx, database.CreateProductToAttributeValueParams{
			ProductID: uuid.NullUUID{
				UUID:  productID,
				Valid: true,
			},
			AttributeValueID: uuid.NullUUID{
				UUID:  attributeValueID,
				Valid: true,
			},
		})

		r.logger.Debug("created product to attribute value", zap.String("productID", productID.String()), zap.String("attributeValueID", attributeValueID.String()))
		return err
	})
	if err != nil {
		r.logger.Error("failed to create product to attribute value", zap.Error(err))
		return err
	}

	r.logger.Info("product to attribute value created", zap.String("productID", productID.String()), zap.String("attributeValueID", attributeValueID.String()))
	return err
}
