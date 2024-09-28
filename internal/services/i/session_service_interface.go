package i

import (
	"context"
	"time"
	"weblineBackend/internal/model"

	"github.com/google/uuid"
)

type SessionService interface {
	CreateSession(ctx context.Context, userID *uuid.UUID, expiresAt time.Time) (model.Session, error)
	GetSessionBySessionID(ctx context.Context, sessionID string) (model.Session, error)
	GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]model.Session, error)
	LinkSessionToUser(ctx context.Context, sessionID string, userID uuid.UUID) error
	DeleteSessionBySessionID(ctx context.Context, sessionID string) error
	UpdateSessionLastActivity(ctx context.Context, sessionID string) error
}
