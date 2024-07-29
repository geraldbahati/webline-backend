package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"weblineBackend/internal/database"
)

type UserRoleRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

// NewUserRoleRepository initializes a new UserRoleRepository with dependency injection for logging
func NewUserRoleRepository(db *sql.DB, logger *zap.Logger) *UserRoleRepository {
	return &UserRoleRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *UserRoleRepository) execTx(ctx context.Context, fn func(*database.Queries) error) (err error) {
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

// AssignRoleToUser adds a role to a user
func (r *UserRoleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		_, err := q.AssignRoleToUser(ctx, database.AssignRoleToUserParams{
			UserID: uuid.NullUUID{
				UUID:  userID,
				Valid: true,
			},
			RoleID: uuid.NullUUID{
				UUID:  roleID,
				Valid: true,
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
	return err
}

// GetRolesForUser returns a list of roles for a user
func (r *UserRoleRepository) GetRolesForUser(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var roles []string
	err := r.execTx(ctx, func(q *database.Queries) error {
		rows, err := q.GetUserRolesByUserID(ctx, uuid.NullUUID{
			UUID:  userID,
			Valid: true,
		})
		if err != nil {
			return err
		}

		for _, row := range rows {
			roles = append(roles, row)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return roles, nil
}
