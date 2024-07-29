package repository

import (
	"context"
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
)

type RoleRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

// NewRoleRepository initializes a new RoleRepository with dependency injection for logging
func NewRoleRepository(db *sql.DB, logger *zap.Logger) *RoleRepository {
	return &RoleRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *RoleRepository) execTx(ctx context.Context, fn func(*database.Queries) error) (err error) {
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

// CreateRole creates a new role
func (r *RoleRepository) CreateRole(ctx context.Context, name, description string) (*model.Role, error) {
	var role *model.Role
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		rows, err := q.CreateRole(ctx, database.CreateRoleParams{
			RoleName: name,
			Description: sql.NullString{
				String: description,
				Valid:  description != "",
			},
		})
		if err != nil {
			return err
		}

		role = &model.Role{
			ID:          rows.ID,
			Name:        rows.RoleName,
			Description: rows.Description.String,
		}
		return err
	})
	if err != nil {
		r.logger.Error("Failed to create role", zap.Error(err))
		return nil, err
	}
	return role, nil
}

// GetRoleByName retrieves a role by name
func (r *RoleRepository) GetRoleByName(ctx context.Context, name string) (*database.GetRoleByNameRow, error) {
	role, err := r.Queries.GetRoleByName(ctx, name)
	if err != nil {
		r.logger.Error("Failed to get role by name", zap.Error(err))
		return nil, err
	}
	return &role, nil
}
