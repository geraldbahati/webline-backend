package sqlc

import (
	"context"
	"database/sql"
	"fmt"
	"weblineBackend/internal/database"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type companyRepositoryImpl struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewCompanyRepositoryImpl(db *sql.DB, logger *zap.Logger) *companyRepositoryImpl {
	return &companyRepositoryImpl{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

func (r *companyRepositoryImpl) execTx(ctx context.Context, fn func(*database.Queries) error) (err error) {
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

func (r *companyRepositoryImpl) CreateCompany(ctx context.Context, name string, kraPIN string, address string, phone string, email string) (*uuid.UUID, error) {
	var id uuid.UUID
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		id, err = q.CreateCompany(ctx, database.CreateCompanyParams{
			Name:   name,
			KraPin: kraPIN,
			Address: sql.NullString{
				String: address,
				Valid:  address != "",
			},
			PhoneNumber: sql.NullString{
				String: phone,
				Valid:  phone != "",
			},
			Email: sql.NullString{
				String: email,
				Valid:  email != "",
			},
		})
		if err != nil {
			r.logger.Error("failed to create company", zap.Error(err))
			return fmt.Errorf("create company: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to create company", zap.Error(err))
		return nil, err
	}
	r.logger.Info("company created successfully", zap.String("id", id.String()))
	return &id, nil
}

func (r *companyRepositoryImpl) GetCompanyID(ctx context.Context, name string, kraPIN string) (*uuid.UUID, error) {
	companyID, err := r.Queries.GetCompany(ctx, database.GetCompanyParams{
		Name:   name,
		KraPin: kraPIN,
	})
	if err != nil {
		r.logger.Error("failed to get company", zap.Error(err))
		return nil, err
	}
	return &companyID, nil
}
