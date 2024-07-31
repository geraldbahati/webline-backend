package model

import (
	"github.com/google/uuid"
	"time"
)

type AdminRequest struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userID"`
	Email     string    `json:"email"`
	Reason    string    `json:"reason"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ApprovalToken struct {
	ID             uuid.UUID `json:"id"`
	Token          string    `json:"token"`
	AdminRequestID uuid.UUID `json:"adminRequestID"`
	ExpiresAt      time.Time `json:"expiresAt"`
}
