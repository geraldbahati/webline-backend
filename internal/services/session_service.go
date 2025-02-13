package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	maxSessionsPerUser     = 5
	sessionGracePeriod     = 30 * 24 * time.Hour // 30 days
	sessionCleanupBatch    = 100
	sessionRefreshInterval = 24 * time.Hour
	minSessionLifetime     = 24 * time.Hour // Minimum session lifetime for permanent login
	sessionDuration        = 24 * time.Hour
	csrfTokenLength        = 32
)

type SessionService struct {
	logger            *zap.Logger
	sessionRepository repository.SessionRepository
	cacheService      CacheService
}

func NewSessionService(logger *zap.Logger, sessionRepository repository.SessionRepository, cacheService CacheService) *SessionService {
	return &SessionService{
		logger:            logger.Named("session_service"),
		sessionRepository: sessionRepository,
		cacheService:      cacheService,
	}
}

// generateCSRFToken generates a cryptographically secure random token
func generateCSRFToken() (string, error) {
	tokenBytes := make([]byte, csrfTokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}

// CreateSession creates a new session with proper CSRF handling
func (s *SessionService) CreateSession(ctx context.Context, userID *uuid.UUID, expiresAt time.Time, csrfToken string) (model.Session, error) {
	s.logger.Debug("Creating new session", zap.Any("userID", userID), zap.Time("expiresAt", expiresAt), zap.String("csrfToken", csrfToken))
	sessionID := uuid.New()

	// Validate expiration time
	if expiresAt.Before(time.Now()) {
		s.logger.Error("invalid expiration time", zap.Time("expiresAt", expiresAt))
		return model.Session{}, fmt.Errorf("invalid expiration time")
	}

	createdSession, err := s.sessionRepository.CreateSession(ctx, userID, sessionID, csrfToken, expiresAt)
	if err != nil {
		s.logger.Error("failed to create session",
			zap.Error(err),
			zap.Any("userID", userID))
		return model.Session{}, fmt.Errorf("session creation failed: %w", err)
	}
	s.logger.Debug("Session created", zap.String("sessionID", createdSession.SessionID.String()), zap.Any("userID", createdSession.UserID))

	// Remove any stale cached value before re-caching the updated session.
	cacheKey := SessionKey(createdSession.SessionID.String())
	_ = s.cacheService.Delete(ctx, cacheKey)

	// Cache the session with the updated values.
	if err := s.cacheSession(ctx, createdSession); err != nil {
		s.logger.Warn("failed to cache session",
			zap.Error(err),
			zap.String("sessionID", sessionID.String()))
	}

	s.logger.Debug("Session created", zap.String("sessionID", createdSession.SessionID.String()))
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

// cacheSession stores a session in the cache
func (s *SessionService) cacheSession(ctx context.Context, session model.Session) error {
	cacheKey := SessionKey(session.SessionID.String())
	expiryDuration := time.Until(session.ExpiresAt)

	if err := s.cacheService.SetWithTTL(ctx, cacheKey, session, expiryDuration); err != nil {
		return fmt.Errorf("cache set: %w", err)
	}

	// If session has a user, add to user's sessions set
	if session.UserID != nil {
		userSessionsKey := UserSessionsKey(session.UserID.String())
		if err := s.cacheService.SAdd(ctx, userSessionsKey, session.SessionID.String()); err != nil {
			s.logger.Warn("Failed to add session to user's set",
				zap.Error(err),
				zap.String("userID", session.UserID.String()))
		}
	}

	return nil
}

// invalidateSessionCache removes a session from all caches
func (s *SessionService) invalidateSessionCache(ctx context.Context, sessionID string, userID *uuid.UUID) {
	// Remove session data
	cacheKey := SessionKey(sessionID)
	if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to delete session from cache",
			zap.Error(err),
			zap.String("sessionID", sessionID))
	}

	// Remove from user's sessions set if applicable
	if userID != nil {
		userSessionsKey := UserSessionsKey(userID.String())
		if err := s.cacheService.SRem(ctx, userSessionsKey, sessionID); err != nil {
			s.logger.Warn("Failed to remove session from user's set",
				zap.Error(err),
				zap.String("userID", userID.String()))
		}
	}

	// Remove expiry tracking
	expiryKey := SessionExpiryKey(sessionID)
	if err := s.cacheService.Delete(ctx, expiryKey); err != nil {
		s.logger.Warn("Failed to delete session expiry",
			zap.Error(err),
			zap.String("sessionID", sessionID))
	}
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
		s.logger.Warn("failed to get session", zap.Error(err), zap.String("sessionID", sessionID))
		return fmt.Errorf("session not found: %w", err)
	}

	// Enhanced session fixation protection
	if session.UserID != nil {
		s.logger.Warn("session already linked to another user", zap.String("sessionID", sessionID), zap.Any("userID", session.UserID))
		return fmt.Errorf("session already linked to another user")
	}

	// if time.Since(session.LastActivity) > sessionGracePeriod {
	// 	s.logger.Debug("session too old to link",
	// 		zap.String("sessionID", sessionID),
	// 		zap.Any("lastActivity", session.LastActivity),
	// 		zap.Duration("sessionGracePeriod", sessionGracePeriod),
	// 		zap.Duration("timeSinceLastActivity", time.Since(session.LastActivity)),
	// 	)
	// 	s.logger.Warn("session too old to link", zap.String("sessionID", sessionID), zap.Any("lastActivity", session.LastActivity))
	// 	return fmt.Errorf("session too old to link")
	// }

	// if session.ExpiresAt.Before(time.Now().Add(minSessionLifetime)) {
	// 	return fmt.Errorf("session expiration too short for permanent login")
	// }

	// Link the session to the user by setting the user id.
	session.UserID = &userID
	// New debug log to confirm session is now linked with the given user id.
	s.logger.Debug("Session linked to user", zap.String("sessionID", sessionID), zap.Any("userID", session.UserID))
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

	// Invalidate the cached value.
	cacheKey := SessionKey(session.SessionID.String())
	_ = s.cacheService.Delete(ctx, cacheKey)

	// Re-cache the session with the updated values.
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

// CreateGuestSession creates a new guest session
func (s *SessionService) CreateGuestSession(ctx context.Context) (*model.Session, error) {
	// Generate session ID and CSRF token
	sessionID := uuid.New()
	csrfToken, err := generateCSRFToken()
	if err != nil {
		s.logger.Error("Failed to generate CSRF token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	// Create session with expiration
	expiresAt := time.Now().Add(sessionDuration)
	session := &model.Session{
		SessionID: sessionID,
		CSRFToken: csrfToken,
		ExpiresAt: expiresAt,
	}

	// Store in repository
	_, err = s.sessionRepository.CreateSession(ctx, nil, sessionID, csrfToken, expiresAt)
	if err != nil {
		s.logger.Error("Failed to create session in repository", zap.Error(err))
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Use the helper method for caching
	if err := s.cacheSession(ctx, *session); err != nil {
		s.logger.Warn("Failed to cache guest session",
			zap.Error(err),
			zap.String("sessionID", sessionID.String()))
	}

	return session, nil
}

// MergeGuestSession merges a guest session with a user session
func (s *SessionService) MergeGuestSession(ctx context.Context, sessionID string, userID uuid.UUID) (*model.Session, error) {
	// Get existing session
	session, err := s.GetSessionBySessionID(ctx, sessionID)
	if err != nil {
		s.logger.Error("Failed to get session", zap.Error(err))
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Generate new CSRF token
	csrfToken, err := generateCSRFToken()
	if err != nil {
		s.logger.Error("Failed to generate CSRF token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	// Update session with user ID and new CSRF token
	session.UserID = &userID
	session.CSRFToken = csrfToken
	session.ExpiresAt = time.Now().Add(sessionDuration)

	// Update in repository
	err = s.sessionRepository.UpdateSession(ctx, session)
	if err != nil {
		s.logger.Error("Failed to update session", zap.Error(err))
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	// Use the helper method for caching
	if err := s.cacheSession(ctx, session); err != nil {
		s.logger.Warn("Failed to cache merged session",
			zap.Error(err),
			zap.String("sessionID", sessionID))
	}

	return &session, nil
}

// DeleteSession deletes a session
func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("parse session ID: %w", err)
	}

	// Get session to get userID before deletion
	session, err := s.GetSessionBySessionID(ctx, sessionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Warn("Failed to get session before deletion",
			zap.Error(err),
			zap.String("sessionID", sessionID))
	}

	// Delete from repository
	err = s.sessionRepository.DeleteSessionBySessionID(ctx, sessionUUID)
	if err != nil {
		s.logger.Error("Failed to delete session from repository",
			zap.Error(err),
			zap.String("sessionID", sessionID))
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Invalidate all cache entries
	s.invalidateSessionCache(ctx, sessionID, session.UserID)

	return nil
}

// Add a new method for cleaning up expired sessions
func (s *SessionService) CleanupExpiredSessions(ctx context.Context) error {
	pattern := fmt.Sprintf("%s:*", NamespaceSession)
	keys, err := s.cacheService.Keys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("get session keys: %w", err)
	}

	now := time.Now()
	for _, key := range keys {
		var session model.Session
		if err := s.cacheService.Get(ctx, key, &session); err != nil {
			continue
		}

		if session.ExpiresAt.Before(now) {
			s.DeleteSession(ctx, session.SessionID.String())
		}
	}

	return nil
}
