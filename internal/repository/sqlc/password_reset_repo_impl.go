package sqlc

import (
	"context"
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"time"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
)

type PasswordResetRepositoryImpl struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewPasswordResetRepositoryImpl(db *sql.DB, logger *zap.Logger) *PasswordResetRepositoryImpl {
	return &PasswordResetRepositoryImpl{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *PasswordResetRepositoryImpl) execTx(ctx context.Context, fn func(*database.Queries) error) (err error) {
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

// StorePasswordResetToken stores a password reset token
func (r *PasswordResetRepositoryImpl) StorePasswordResetToken(ctx context.Context, email string, token string, expiresAt time.Time) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.StorePasswordResetToken(ctx, database.StorePasswordResetTokenParams{
			Email:     email,
			Token:     token,
			ExpiresAt: expiresAt,
		})
		return err
	})
	return err
}

// GetPasswordResetToken gets a password reset token
func (r *PasswordResetRepositoryImpl) GetPasswordResetToken(ctx context.Context, email string) (*model.PasswordReset, error) {
	var passwordReset *model.PasswordReset
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		row, err := q.GetPasswordResetToken(ctx, email)
		if err != nil {
			r.logger.Error("get password reset token", zap.Error(err))
			return fmt.Errorf("get password reset token: %w", err)
		}

		passwordReset = &model.PasswordReset{
			Email:     row.Email,
			Token:     row.Token,
			ExpiresAt: row.ExpiresAt,
		}

		return nil
	})
	return passwordReset, err
}

// DeletePasswordResetToken deletes a password reset token
func (r *PasswordResetRepositoryImpl) DeletePasswordResetToken(ctx context.Context, email string) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.DeletePasswordResetToken(ctx, email)
		return err
	})
	return err
}
