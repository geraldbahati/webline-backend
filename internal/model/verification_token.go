package model

import (
	"github.com/google/uuid"
	"time"
)

type VerificationToken struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}
