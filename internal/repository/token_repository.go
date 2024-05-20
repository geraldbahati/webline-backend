package repository

import (
	"context"
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"weblineBackend/internal/database"
	"weblineBackend/pkg/logger"
)

type TokenRepository struct {
	*database.Queries
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{
		Queries: database.New(db),
		db:      db,
	}
}

func (r *TokenRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	q := database.New(tx)
	if err := fn(q); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback transaction: %w", rbErr)
		}
		return fmt.Errorf("exec transaction: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// StoreRefreshToken stores a refresh token
func (r *TokenRepository) StoreRefreshToken(
	ctx context.Context,
	refreshToken database.StoreRefreshTokenParams,
) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		if _, err := q.StoreRefreshToken(ctx, refreshToken); err != nil {
			return fmt.Errorf("failed to store refresh token: %w", err)
		}
		return nil
	})
	if err != nil {
		logger.Error("failed to store refresh token: ", zap.Error(err))
		return err
	}

	logger.Info("refresh token stored successfully")
	return nil
}
