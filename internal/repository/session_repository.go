package repository

import (
	"context"
	"time"
	"weblineBackend/internal/model"

	"github.com/google/uuid"
)

type SessionRepository interface {
	// CreateSession creates a new session and returns the created session details
	CreateSession(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, csrfToken string, expiresAt time.Time) (model.Session, error)

	// GetSessionBySessionID retrieves a session by its session ID
	GetSessionBySessionID(ctx context.Context, sessionID uuid.UUID) (model.Session, error)

	// GetSessionByUserID retrieves all sessions for a given user ID
	GetSessionByUserID(ctx context.Context, userID uuid.UUID) ([]model.Session, error)

	// LinkSessionToUser links a session to a user and updates the last activity
	LinkSessionToUser(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) error

	// DeleteSessionBySessionID deletes a session by its session ID
	DeleteSessionBySessionID(ctx context.Context, sessionID uuid.UUID) error

	// UpdateSessionLastActivity updates the last activity timestamp for a session
	UpdateSessionLastActivity(ctx context.Context, sessionID uuid.UUID) error
}
