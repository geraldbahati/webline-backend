package sqlc

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"weblineBackend/internal/database"

	"go.uber.org/zap"
)

type settingsRepositoryImpl struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewSettingsRepositoryImpl(db *sql.DB, logger *zap.Logger) *settingsRepositoryImpl {
	return &settingsRepositoryImpl{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

func (r *settingsRepositoryImpl) execTx(ctx context.Context, fn func(*database.Queries) error) (err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return fmt.Errorf("begin transaction: %w", err)
	}

	q := database.New(tx)
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			r.logger.Panic("transaction panicked, rolling back", zap.Any("panic", p))
			panic(p) // Re-throw panic after rollback
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				r.logger.Error("rollback failed", zap.Error(rbErr))
				err = fmt.Errorf("transaction rollback failed: %v, original error: %w", rbErr, err)
			} else {
				r.logger.Warn("transaction rolled back due to error", zap.Error(err))
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				r.logger.Error("commit failed", zap.Error(commitErr))
				err = fmt.Errorf("commit transaction: %w", commitErr)
			}
		}
	}()

	err = fn(q)
	return err
}

func (r *settingsRepositoryImpl) GetVATPercentage(ctx context.Context) (float64, error) {
	vatPercentage, err := r.Queries.GetVATPercentage(ctx)
	if err != nil {
		r.logger.Error("failed to get vat percentage", zap.Error(err))
		return 0, fmt.Errorf("get vat percentage: %w", err)
	}

	vatPercentageFloat, err := strconv.ParseFloat(vatPercentage, 64)
	if err != nil {
		r.logger.Error("failed to convert vat percentage to float", zap.Error(err))
		return 0, fmt.Errorf("convert vat percentage to float: %w", err)
	}

	return vatPercentageFloat, nil
}

func (r *settingsRepositoryImpl) UpdateVATPercentage(ctx context.Context, percentage float64) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.UpdateVATPercentage(ctx, fmt.Sprintf("%f", percentage))
		if err != nil {
			r.logger.Error("failed to update vat percentage", zap.Error(err))
			return fmt.Errorf("update vat percentage: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to update vat percentage", zap.Error(err))
		return fmt.Errorf("update vat percentage: %w", err)
	}
	return nil
}
