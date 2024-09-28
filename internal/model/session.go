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
