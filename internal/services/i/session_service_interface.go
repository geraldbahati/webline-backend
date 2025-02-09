package i

import (
	"context"
	"time"
	"weblineBackend/internal/model"

	"github.com/google/uuid"
)

type SessionService interface {
	// CreateSession returns a session which now includes the user_id field
	// if the session is linked to a user.
	CreateSession(ctx context.Context, userID *uuid.UUID, expiresAt time.Time, csrfToken string) (model.Session, error)
	GetSessionBySessionID(ctx context.Context, sessionID string) (model.Session, error)
	GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]model.Session, error)
	LinkSessionToUser(ctx context.Context, sessionID string, userID uuid.UUID) error
	DeleteSessionBySessionID(ctx context.Context, sessionID string) error
	UpdateSession(ctx context.Context, session model.Session) error
	InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error
}
