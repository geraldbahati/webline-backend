package model

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID           uuid.UUID  `json:"id"`
	SessionID    uuid.UUID  `json:"sessionID"`
	UserID       *uuid.UUID `json:"userID,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	LastActivity time.Time  `json:"lastActivity"`
	ExpiresAt    time.Time  `json:"expiresAt"`
	CSRFToken    string     `json:"csrfToken"`
}

// IsExpired returns true if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Refresh updates the session's LastActivity and ExpiresAt values.
func (s *Session) Refresh(newExpiry time.Time) {
	s.LastActivity = time.Now()
	s.ExpiresAt = newExpiry
}
