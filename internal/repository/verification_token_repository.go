package repository

import (
	"context"
	"time"
	"weblineBackend/internal/model"
)

type VerificationTokenRepository interface {
	CreateVerificationToken(ctx context.Context, email string, token string, expiresAt time.Time) (*model.VerificationToken, error)
	GetVerificationTokenByEmail(ctx context.Context, email string) (*model.VerificationToken, error)
	GetVerificationTokenByToken(ctx context.Context, token string) (*model.VerificationToken, error)
	DeleteVerificationToken(ctx context.Context, token string) error
	DeleteVerificationTokens(ctx context.Context, email string) error
}
