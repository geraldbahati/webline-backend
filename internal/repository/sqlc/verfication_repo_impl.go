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

type VerificationTokenRepositoryImpl struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewVerificationTokenRepositoryImpl(db *sql.DB, logger *zap.Logger) *VerificationTokenRepositoryImpl {
	return &VerificationTokenRepositoryImpl{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *VerificationTokenRepositoryImpl) execTx(ctx context.Context, fn func(*database.Queries) error) (err error) {
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

// CreateVerificationToken creates a new verification token
func (r *VerificationTokenRepositoryImpl) CreateVerificationToken(ctx context.Context, email string, token string, expiresAt time.Time) (*model.VerificationToken, error) {
	var verificationToken *model.VerificationToken
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		row, err := q.CreateVerificationToken(ctx, database.CreateVerificationTokenParams{
			Email:     email,
			Token:     token,
			ExpiresAt: expiresAt,
		})
		if err != nil {
			r.logger.Error("create verification token", zap.Error(err))
			return fmt.Errorf("create verification token: %w", err)
		}

		verificationToken = &model.VerificationToken{
			ID:        row.ID,
			Email:     row.Email,
			Token:     row.Token,
			ExpiresAt: row.ExpiresAt,
		}

		return nil
	})
	if err != nil {
		r.logger.Error("create verification token", zap.Error(err))
		return nil, fmt.Errorf("create verification token: %w", err)
	}
	return verificationToken, nil
}

// GetVerificationTokenByEmail retrieves a verification token by email
func (r *VerificationTokenRepositoryImpl) GetVerificationTokenByEmail(ctx context.Context, email string) (*model.VerificationToken, error) {
	row, err := r.Queries.GetVerificationTokenByEmail(ctx, email)
	if err != nil {
		r.logger.Error("get verification token by email", zap.Error(err))
		return nil, fmt.Errorf("get verification token by email: %w", err)
	}

	verificationToken := &model.VerificationToken{
		ID:        row.ID,
		Email:     row.Email,
		Token:     row.Token,
		ExpiresAt: row.ExpiresAt,
	}

	return verificationToken, nil
}

// DeleteVerificationToken deletes a verification token by email
func (r *VerificationTokenRepositoryImpl) DeleteVerificationToken(ctx context.Context, token string) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.DeleteVerificationTokenByToken(ctx, token)
		if err != nil {
			r.logger.Error("delete verification token", zap.Error(err))
			return fmt.Errorf("delete verification token: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("delete verification token", zap.Error(err))
		return fmt.Errorf("delete verification token: %w", err)
	}
	return nil
}

// GetVerificationTokenByToken retrieves a verification token by token
func (r *VerificationTokenRepositoryImpl) GetVerificationTokenByToken(ctx context.Context, token string) (*model.VerificationToken, error) {
	row, err := r.Queries.GetVerificationTokenByToken(ctx, token)
	if err != nil {
		r.logger.Error("get verification token by token", zap.Error(err))
		return nil, fmt.Errorf("get verification token by token: %w", err)
	}

	verificationToken := &model.VerificationToken{
		ID:        row.ID,
		Email:     row.Email,
		Token:     row.Token,
		ExpiresAt: row.ExpiresAt,
	}

	return verificationToken, nil
}

// DeleteVerificationTokens deletes all verification tokens by email
func (r *VerificationTokenRepositoryImpl) DeleteVerificationTokens(ctx context.Context, email string) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.DeleteVerificationTokensByEmail(ctx, email)
		if err != nil {
			r.logger.Error("delete verification tokens", zap.Error(err))
			return fmt.Errorf("delete verification tokens: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("delete verification tokens", zap.Error(err))
		return fmt.Errorf("delete verification tokens: %w", err)
	}
	return nil
}
