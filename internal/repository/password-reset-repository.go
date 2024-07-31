package repository

import (
	"context"
	"time"
	"weblineBackend/internal/model"
)

type PasswordResetRepository interface {
	StorePasswordResetToken(ctx context.Context, email string, token string, expiresAt time.Time) error
	GetPasswordResetToken(ctx context.Context, email string) (*model.PasswordReset, error)
	DeletePasswordResetToken(ctx context.Context, email string) error
}
