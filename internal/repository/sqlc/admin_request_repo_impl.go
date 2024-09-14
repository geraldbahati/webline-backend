package sqlc

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AdminRequestRepositoryImpl struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewAdminRequestRepositoryImpl(db *sql.DB, logger *zap.Logger) *AdminRequestRepositoryImpl {
	return &AdminRequestRepositoryImpl{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *AdminRequestRepositoryImpl) execTx(ctx context.Context, fn func(*database.Queries) error) (err error) {
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

// CreateAdminRequest creates a new admin request
func (r *AdminRequestRepositoryImpl) CreateAdminRequest(ctx context.Context, userID uuid.UUID, reason string) (uuid.UUID, error) {
	var id uuid.UUID

	err := r.execTx(ctx, func(q *database.Queries) error {
		questID, err := q.CreateAdminRequest(ctx, database.CreateAdminRequestParams{
			UserID: userID,
			Reason: reason,
		})
		if err != nil {
			return err
		}

		id = questID
		return err
	})
	return id, err
}

// GetPendingAdminRequests returns all pending admin requests
func (r *AdminRequestRepositoryImpl) GetPendingAdminRequests(ctx context.Context) ([]model.AdminRequest, error) {
	adminRequests, err := r.Queries.GetPendingAdminRequests(ctx)
	if err != nil {
		return nil, err
	}

	var result []model.AdminRequest
	for _, adminRequest := range adminRequests {
		result = append(result, model.AdminRequest{
			ID:        adminRequest.ID,
			UserID:    adminRequest.UserID,
			Email:     adminRequest.Email,
			Reason:    adminRequest.Reason,
			Status:    adminRequest.Status,
			CreatedAt: adminRequest.CreatedAt,
			UpdatedAt: adminRequest.UpdatedAt,
		})
	}

	return result, nil
}

// GetAdminRequestByID returns an admin request by ID
func (r *AdminRequestRepositoryImpl) GetAdminRequestByID(ctx context.Context, id uuid.UUID) (model.AdminRequest, error) {
	adminRequest, err := r.Queries.GetAdminRequestByID(ctx, id)
	if err != nil {
		return model.AdminRequest{}, err
	}

	return model.AdminRequest{
		ID:        adminRequest.ID,
		UserID:    adminRequest.UserID,
		Email:     adminRequest.Email,
		Reason:    adminRequest.Reason,
		Status:    adminRequest.Status,
		CreatedAt: adminRequest.CreatedAt,
		UpdatedAt: adminRequest.UpdatedAt,
	}, nil
}

// ApproveAdminRequest approves an admin request
func (r *AdminRequestRepositoryImpl) ApproveAdminRequest(ctx context.Context, id uuid.UUID) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.ApproveAdminRequest(ctx, id)
		return err
	})
	return err
}

// RejectAdminRequest rejects an admin request
func (r *AdminRequestRepositoryImpl) RejectAdminRequest(ctx context.Context, id uuid.UUID) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.RejectAdminRequest(ctx, id)
		return err
	})
	return err
}

// GetAdminRequestsByUserID returns all admin requests by user ID
func (r *AdminRequestRepositoryImpl) GetAdminRequestsByUserID(ctx context.Context, userID uuid.UUID) ([]model.AdminRequest, error) {
	adminRequests, err := r.Queries.GetAdminRequestsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []model.AdminRequest
	for _, adminRequest := range adminRequests {
		result = append(result, model.AdminRequest{
			ID:        adminRequest.ID,
			UserID:    adminRequest.UserID,
			Email:     adminRequest.Email,
			Reason:    adminRequest.Reason,
			Status:    adminRequest.Status,
			CreatedAt: adminRequest.CreatedAt,
			UpdatedAt: adminRequest.UpdatedAt,
		})
	}

	return result, nil
}

// StoreApprovalToken stores an approval token
func (r *AdminRequestRepositoryImpl) StoreApprovalToken(ctx context.Context, token string, adminRequestID uuid.UUID, expiresAt time.Time) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.StoreApprovalToken(ctx, database.StoreApprovalTokenParams{
			Token:     token,
			RequestID: adminRequestID,
			ExpiresAt: expiresAt,
		})

		return err
	})
	return err
}

// GetApprovalToken returns an approval token by token
func (r *AdminRequestRepositoryImpl) GetApprovalToken(ctx context.Context, token string) (model.ApprovalToken, error) {
	approvalToken, err := r.Queries.GetApprovalToken(ctx, token)
	if err != nil {
		return model.ApprovalToken{}, err
	}

	return model.ApprovalToken{
		ID:             approvalToken.ID,
		Token:          approvalToken.Token,
		AdminRequestID: approvalToken.RequestID,
		ExpiresAt:      approvalToken.ExpiresAt,
	}, nil
}

// DeleteApprovalToken deletes an approval token by token
func (r *AdminRequestRepositoryImpl) DeleteApprovalToken(ctx context.Context, token string) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.DeleteApprovalToken(ctx, token)
		return err
	})
	return err
}

// GetAdminRequestByUserID returns an admin request by user ID
func (r *AdminRequestRepositoryImpl) GetAdminRequestByUserID(ctx context.Context, userID uuid.UUID) (model.AdminRequest, error) {
	adminRequest, err := r.Queries.GetAdminRequestByUserID(ctx, userID)
	if err != nil {
		return model.AdminRequest{}, err
	}

	return model.AdminRequest{
		ID:        adminRequest.ID,
		UserID:    adminRequest.UserID,
		Reason:    adminRequest.Reason,
		Status:    adminRequest.Status,
		CreatedAt: adminRequest.CreatedAt,
		UpdatedAt: adminRequest.UpdatedAt,
	}, nil
}
