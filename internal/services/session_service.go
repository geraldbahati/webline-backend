package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
	"weblineBackend/internal/services/i"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SessionService struct {
	logger            *zap.Logger
	sessionRepository repository.SessionRepository
	cacheService      CacheService
}

func NewSessionService(logger *zap.Logger, sessionRepository repository.SessionRepository, cacheService CacheService) i.SessionService {
	return &SessionService{
		logger:            logger,
		sessionRepository: sessionRepository,
		cacheService:      cacheService,
	}
}

// generateCSRFToken generates a secure random token
func generateCSRFToken() (string, error) {
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenBytes), nil
}

// CreateSession creates a new session and caches it
func (s *SessionService) CreateSession(ctx context.Context, userID *uuid.UUID, expiresAt time.Time) (model.Session, error) {
	sessionID := uuid.New()
	csrfToken, err := generateCSRFToken()
	if err != nil {
		s.logger.Error("Failed to generate CSRF token", zap.Error(err))
		return model.Session{}, fmt.Errorf("generate CSRF token: %w", err)
	}

	session, err := s.sessionRepository.CreateSession(ctx, userID, sessionID, csrfToken, expiresAt)
	if err != nil {
		s.logger.Error("Failed to create session", zap.Error(err))
		return model.Session{}, fmt.Errorf("create session: %w", err)
	}

	// Cache the session
	cacheKey := SessionKey(sessionID.String())
	err = s.cacheService.Set(ctx, cacheKey, session)
	if err != nil {
		s.logger.Warn("Failed to cache session", zap.Error(err))
	}

	return session, nil
}

// GetSessionBySessionID retrieves a session by its ID, using cache if available
func (s *SessionService) GetSessionBySessionID(ctx context.Context, sessionID string) (model.Session, error) {
	var session model.Session
	cacheKey := SessionKey(sessionID)

	err := s.cacheService.GetOrSet(ctx, cacheKey, &session, func() error {
		sessionID, err := uuid.Parse(sessionID)
		if err != nil {
			return fmt.Errorf("parse session ID: %w", err)
		}

		dbSession, err := s.sessionRepository.GetSessionBySessionID(ctx, sessionID)
		if err != nil {
			return err
		}
		session = dbSession
		return nil
	})

	if err != nil {
		s.logger.Error("Failed to get session by ID", zap.Error(err))
		return model.Session{}, fmt.Errorf("get session by ID: %w", err)
	}

	return session, nil
}

// Update SessionService.GetSessionsByUserID to accept uuid.UUID
func (s *SessionService) GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]model.Session, error) {
	sessions, err := s.sessionRepository.GetSessionByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get sessions by user ID", zap.Error(err))
		return nil, fmt.Errorf("get sessions by user ID: %w", err)
	}

	return sessions, nil
}

func (s *SessionService) LinkSessionToUser(ctx context.Context, sessionID string, userID uuid.UUID) error {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("parse session ID: %w", err)
	}

	err = s.sessionRepository.LinkSessionToUser(ctx, sessionUUID, userID)
	if err != nil {
		s.logger.Error("Failed to link session to user", zap.Error(err))
		return fmt.Errorf("link session to user: %w", err)
	}

	// Update the cached session
	session, err := s.GetSessionBySessionID(ctx, sessionID)
	if err != nil {
		s.logger.Warn("Failed to get updated session for cache", zap.Error(err))
	} else {
		cacheKey := SessionKey(sessionID)
		err = s.cacheService.Set(ctx, cacheKey, session)
		if err != nil {
			s.logger.Warn("Failed to update cached session", zap.Error(err))
		}
	}

	return nil
}

// DeleteSessionBySessionID deletes a session and removes it from cache
func (s *SessionService) DeleteSessionBySessionID(ctx context.Context, sessionID string) error {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("parse session ID: %w", err)
	}

	err = s.sessionRepository.DeleteSessionBySessionID(ctx, sessionUUID)
	if err != nil {
		s.logger.Error("Failed to delete session", zap.Error(err))
		return fmt.Errorf("delete session: %w", err)
	}

	// Remove from cache
	cacheKey := SessionKey(sessionID)
	err = s.cacheService.Delete(ctx, cacheKey)
	if err != nil {
		s.logger.Warn("Failed to remove session from cache", zap.Error(err))
	}

	return nil
}

// UpdateSessionLastActivity updates the last activity timestamp for a session and updates the cache
func (s *SessionService) UpdateSessionLastActivity(ctx context.Context, sessionID string) error {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("parse session ID: %w", err)
	}

	err = s.sessionRepository.UpdateSessionLastActivity(ctx, sessionUUID)
	if err != nil {
		s.logger.Error("Failed to update session last activity", zap.Error(err))
		return fmt.Errorf("update session last activity: %w", err)
	}

	// Update the cached session
	session, err := s.GetSessionBySessionID(ctx, sessionID)
	if err != nil {
		s.logger.Warn("Failed to get updated session for cache", zap.Error(err))
	} else {
		cacheKey := SessionKey(sessionID)
		err = s.cacheService.Set(ctx, cacheKey, session)
		if err != nil {
			s.logger.Warn("Failed to update cached session", zap.Error(err))
		}
	}

	return nil
}
