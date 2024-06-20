package repository

import (
	"context"
	"database/sql"
	"fmt"
	"weblineBackend/internal/database"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type GuestCheckoutRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewGuestCheckoutRepository(db *sql.DB, logger *zap.Logger) *GuestCheckoutRepository {
	return &GuestCheckoutRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *GuestCheckoutRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
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

// CreateGuestCheckout creates a new guest checkout record
func (r *GuestCheckoutRepository) CreateGuestCheckout(ctx context.Context, params *database.CreateGuestCheckoutParams) (*uuid.UUID, error) {
	var guestID uuid.UUID
	if err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		guestID, err = q.CreateGuestCheckout(ctx, *params)
		if err != nil {
			r.logger.Error("create guest checkout failed", zap.Error(err))
			return fmt.Errorf("create guest checkout: %w", err)
		}
		return nil
	}); err != nil {
		r.logger.Error("create guest checkout transaction failed", zap.Error(err))
		return nil, fmt.Errorf("create guest checkout transaction: %w", err)
	}

	r.logger.Info("guest checkout created", zap.String("guestID", guestID.String()))
	return &guestID, nil
}

// GetGuestCheckoutByEmail retrieves a guest checkout record by email
func (r *GuestCheckoutRepository) GetGuestCheckoutByEmail(ctx context.Context, email string) (*database.GuestCheckout, error) {
	guest, err := r.Queries.GetGuestCheckoutByEmail(ctx, email)
	if err != nil {
		r.logger.Error("get guest checkout by email failed", zap.Error(err))
		return nil, fmt.Errorf("get guest checkout by email: %w", err)
	}

	return &guest, nil
}
