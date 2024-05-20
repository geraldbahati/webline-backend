package utils

import (
	"context"

	"github.com/google/uuid"
)

func SetUserIdInContext(ctx context.Context, userId uuid.UUID) context.Context {
	return context.WithValue(ctx, "userId", userId)
}
