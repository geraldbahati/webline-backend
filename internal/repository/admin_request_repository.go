package repository

import (
	"context"
	"github.com/google/uuid"
	"time"
	"weblineBackend/internal/model"
)

type AdminRequestRepository interface {
	CreateAdminRequest(ctx context.Context, userID uuid.UUID, reason string) (uuid.UUID, error)
	GetPendingAdminRequests(ctx context.Context) ([]model.AdminRequest, error)
	GetAdminRequestByID(ctx context.Context, id uuid.UUID) (model.AdminRequest, error)
	ApproveAdminRequest(ctx context.Context, id uuid.UUID) error
	RejectAdminRequest(ctx context.Context, id uuid.UUID) error
	GetAdminRequestsByUserID(ctx context.Context, userID uuid.UUID) ([]model.AdminRequest, error)
	StoreApprovalToken(ctx context.Context, token string, adminRequestID uuid.UUID, expiresAt time.Time) error
	GetApprovalToken(ctx context.Context, token string) (model.ApprovalToken, error)
	DeleteApprovalToken(ctx context.Context, token string) error
}
