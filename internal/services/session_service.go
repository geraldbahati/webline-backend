package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
	"weblineBackend/internal/services/i"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	maxSessionsPerUser     = 5
	sessionGracePeriod     = 5 * time.Minute
	sessionCleanupBatch    = 100
	sessionRefreshInterval = 24 * time.Hour
	minSessionLifetime     = 24 * time.Hour // Minimum session lifetime for permanent login
)

type SessionService struct {
	logger            *zap.Logger
	sessionRepository repository.SessionRepository
	cacheService      CacheService
}

func NewSessionService(logger *zap.Logger, sessionRepository repository.SessionRepository, cacheService CacheService) i.SessionService {
	return &SessionService{
		logger:            logger.Named("session_service"),
		sessionRepository: sessionRepository,
		cacheService:      cacheService,
	}
}

// generateCSRFToken generates a cryptographically secure random token
func generateCSRFToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}

// CreateSession creates a new session with proper CSRF handling
func (s *SessionService) CreateSession(ctx context.Context, userID *uuid.UUID, expiresAt time.Time, csrfToken string) (model.Session, error) {
	sessionID := uuid.New()

	// Validate expiration time
	if expiresAt.Before(time.Now()) {
		return model.Session{}, fmt.Errorf("invalid expiration time")
	}

	createdSession, err := s.sessionRepository.CreateSession(ctx, userID, sessionID, csrfToken, expiresAt)
	if err != nil {
		s.logger.Error("failed to create session",
			zap.Error(err),
			zap.Any("userID", userID))
		return model.Session{}, fmt.Errorf("session creation failed: %w", err)
	}

	// Cache the session
	if err := s.cacheSession(ctx, createdSession); err != nil {
		s.logger.Warn("failed to cache session",
			zap.Error(err),
			zap.String("sessionID", sessionID.String()))
	}

	return createdSession, nil
}

// GetSessionBySessionID retrieves a session by its ID with caching
func (s *SessionService) GetSessionBySessionID(ctx context.Context, sessionID string) (model.Session, error) {
	var session model.Session
	cacheKey := SessionKey(sessionID)

	// Try cache first
	if err := s.cacheService.Get(ctx, cacheKey, &session); err == nil {
		// Validate session expiration
		if session.ExpiresAt.Before(time.Now()) {
			s.DeleteSessionBySessionID(ctx, sessionID)
			return model.Session{}, fmt.Errorf("session expired")
		}
		return session, nil
	}

	// Fallback to database
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return model.Session{}, fmt.Errorf("invalid session ID format: %w", err)
	}

	session, err = s.sessionRepository.GetSessionBySessionID(ctx, sessionUUID)
	if err != nil {
		return model.Session{}, fmt.Errorf("session not found: %w", err)
	}

	// Validate and cache session
	if session.ExpiresAt.Before(time.Now()) {
		s.DeleteSessionBySessionID(ctx, sessionID)
		return model.Session{}, fmt.Errorf("session expired")
	}

	if err := s.cacheSession(ctx, session); err != nil {
		s.logger.Warn("failed to refresh session cache",
			zap.Error(err),
			zap.String("sessionID", sessionID))
	}

	return session, nil
}

// cacheSession helper function to cache a session
func (s *SessionService) cacheSession(ctx context.Context, session model.Session) error {
	cacheKey := SessionKey(session.SessionID.String())
	ttl := time.Until(session.ExpiresAt)
	return s.cacheService.SetWithTTL(ctx, cacheKey, session, ttl)
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

// findOldestSession returns the session with the oldest LastActivity
func findOldestSession(sessions []model.Session) model.Session {
	if len(sessions) == 0 {
		return model.Session{}
	}

	oldest := sessions[0]
	for _, session := range sessions[1:] {
		if session.LastActivity.Before(oldest.LastActivity) {
			oldest = session
		}
	}
	return oldest
}

// LinkSessionToUser with security checks and session limit enforcement
func (s *SessionService) LinkSessionToUser(ctx context.Context, sessionID string, userID uuid.UUID) error {
	// Check existing sessions
	sessions, err := s.GetSessionsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Clean expired sessions first
	sessions = s.filterExpiredSessions(sessions)

	// Enforce maximum sessions per user
	if len(sessions) >= maxSessionsPerUser {
		oldest := findOldestSession(sessions)
		if err := s.DeleteSessionBySessionID(ctx, oldest.SessionID.String()); err != nil {
			s.logger.Warn("failed to clean up old session",
				zap.Error(err),
				zap.String("sessionID", oldest.SessionID.String()))
		}
	}

	// Get and validate existing session
	session, err := s.GetSessionBySessionID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Enhanced session fixation protection
	if session.UserID != nil {
		return fmt.Errorf("session already linked to another user")
	}

	if time.Since(session.LastActivity) > sessionGracePeriod {
		return fmt.Errorf("session too old to link")
	}

	if session.ExpiresAt.Before(time.Now().Add(minSessionLifetime)) {
		return fmt.Errorf("session expiration too short for permanent login")
	}

	session.UserID = &userID
	return s.UpdateSession(ctx, session)
}

// filterExpiredSessions removes expired sessions from the slice
func (s *SessionService) filterExpiredSessions(sessions []model.Session) []model.Session {
	valid := sessions[:0]
	now := time.Now()

	for _, session := range sessions {
		if session.ExpiresAt.After(now) {
			valid = append(valid, session)
		}
	}
	return valid
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

// UpdateSession updates all session fields with proper cache handling
func (s *SessionService) UpdateSession(ctx context.Context, session model.Session) error {
	if err := s.sessionRepository.UpdateSession(ctx, session); err != nil {
		s.logger.Error("failed to update session",
			zap.Error(err),
			zap.String("sessionID", session.SessionID.String()))
		return fmt.Errorf("session update failed: %w", err)
	}

	// Update cache with proper TTL
	if err := s.cacheSession(ctx, session); err != nil {
		s.logger.Warn("failed to update cached session",
			zap.Error(err),
			zap.String("sessionID", session.SessionID.String()))
	}

	return nil
}

// InvalidateUserSessions logs out all user sessions
func (s *SessionService) InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error {
	sessions, err := s.GetSessionsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	for _, session := range sessions {
		if err := s.DeleteSessionBySessionID(ctx, session.SessionID.String()); err != nil {
			s.logger.Warn("Failed to delete user session",
				zap.Error(err),
				zap.String("sessionID", session.SessionID.String()))
		}
	}

	return nil
}

// RefreshSession extends the session lifetime if it's within the refresh window
func (s *SessionService) RefreshSession(ctx context.Context, sessionID string) (model.Session, error) {
	session, err := s.GetSessionBySessionID(ctx, sessionID)
	if err != nil {
		return model.Session{}, fmt.Errorf("failed to get session: %w", err)
	}

	// Only refresh if within refresh window
	if time.Until(session.ExpiresAt) > sessionRefreshInterval {
		return session, nil
	}

	newExpiry := time.Now().Add(sessionRefreshInterval * 2)
	session.ExpiresAt = newExpiry

	if err := s.UpdateSession(ctx, session); err != nil {
		return model.Session{}, fmt.Errorf("failed to refresh session: %w", err)
	}

	return session, nil
}
