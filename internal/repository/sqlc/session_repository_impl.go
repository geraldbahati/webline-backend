package sqlc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"weblineBackend/internal/app_errors"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type sessionRepositoryImpl struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewSessionRepositoryImpl(db *sql.DB, logger *zap.Logger) repository.SessionRepository {
	return &sessionRepositoryImpl{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

func (r *sessionRepositoryImpl) CreateSession(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, csrfToken string, expiresAt time.Time) (model.Session, error) {
	var userUUID uuid.NullUUID
	if userID != nil {
		userUUID = uuid.NullUUID{UUID: *userID, Valid: true}
	}

	dbSession, err := r.Queries.CreateSession(ctx, database.CreateSessionParams{
		UserID:    userUUID,
		SessionID: sessionID,
		ExpiresAt: expiresAt,
		CsrfToken: csrfToken,
	})
	if err != nil {
		r.logger.Error("failed to create session", zap.Error(err))
		return model.Session{}, fmt.Errorf("create session: %w", err)
	}

	return convertDBSessionToModelSession(dbSession), nil
}

func (r *sessionRepositoryImpl) GetSessionBySessionID(ctx context.Context, sessionID uuid.UUID) (model.Session, error) {
	dbSession, err := r.Queries.GetSessionBySessionID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Session{}, app_errors.NewSessionNotFoundError()
		}
		r.logger.Error("failed to get session by session ID", zap.Error(err))
		return model.Session{}, fmt.Errorf("get session by session ID: %w", err)
	}

	return convertDBSessionToModelSession(dbSession), nil
}

func (r *sessionRepositoryImpl) GetSessionByUserID(ctx context.Context, userID uuid.UUID) ([]model.Session, error) {
	dbSessions, err := r.Queries.GetSessionByUserID(ctx, uuid.NullUUID{UUID: userID, Valid: true})
	if err != nil {
		r.logger.Error("failed to get sessions by user ID", zap.Error(err))
		return nil, fmt.Errorf("get sessions by user ID: %w", err)
	}

	sessions := make([]model.Session, len(dbSessions))
	for i, dbSession := range dbSessions {
		sessions[i] = convertDBSessionToModelSession(dbSession)
	}

	return sessions, nil
}

func (r *sessionRepositoryImpl) LinkSessionToUser(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) error {
	err := r.Queries.LinkSessionToUser(ctx, database.LinkSessionToUserParams{
		SessionID: sessionID,
		UserID:    uuid.NullUUID{UUID: userID, Valid: true},
	})
	if err != nil {
		r.logger.Error("failed to link session to user", zap.Error(err))
		return fmt.Errorf("link session to user: %w", err)
	}

	return nil
}

func (r *sessionRepositoryImpl) DeleteSessionBySessionID(ctx context.Context, sessionID uuid.UUID) error {
	err := r.Queries.DeleteSessionBySessionID(ctx, sessionID)
	if err != nil {
		r.logger.Error("failed to delete session by session ID", zap.Error(err))
		return fmt.Errorf("delete session by session ID: %w", err)
	}

	return nil
}

func (r *sessionRepositoryImpl) UpdateSessionLastActivity(ctx context.Context, sessionID uuid.UUID) error {
	err := r.Queries.UpdateSessionLastActivity(ctx, sessionID)
	if err != nil {
		r.logger.Error("failed to update session last activity", zap.Error(err))
		return fmt.Errorf("update session last activity: %w", err)
	}

	return nil
}

func convertDBSessionToModelSession(dbSession database.Session) model.Session {
	var userID *uuid.UUID
	if dbSession.UserID.Valid {
		userID = &dbSession.UserID.UUID
	}

	return model.Session{
		ID:           dbSession.ID,
		SessionID:    dbSession.SessionID,
		UserID:       userID,
		CreatedAt:    dbSession.CreatedAt,
		LastActivity: dbSession.LastActivity,
		ExpiresAt:    dbSession.ExpiresAt,
		CSRFToken:    dbSession.CsrfToken,
	}
}
