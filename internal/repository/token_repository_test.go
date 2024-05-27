package repository

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
	"time"
	"weblineBackend/internal/database"
)

func TestStoreRefreshToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewTokenRepository(db, logger)

	ctx := context.Background()
	refreshToken := database.StoreRefreshTokenParams{
		UserID:    uuid.New(),
		Token:     "some-refresh-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	expectedID := uuid.New()
	createdAt := time.Now()

	// Expectations for the SQL queries
	mock.ExpectBegin()
	mock.ExpectQuery("(?s).*INSERT INTO refresh_tokens \\(id, user_id, token, created_at, expires_at\\) VALUES \\(gen_random_uuid\\(\\), \\$1, \\$2, NOW\\(\\), \\$3\\) RETURNING id, user_id, token, created_at, expires_at*").
		WithArgs(refreshToken.UserID, refreshToken.Token, refreshToken.ExpiresAt).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "token", "created_at", "expires_at",
		}).AddRow(expectedID, refreshToken.UserID, refreshToken.Token, createdAt, refreshToken.ExpiresAt))
	mock.ExpectCommit()

	err := repo.StoreRefreshToken(ctx, refreshToken)
	require.NoError(t, err)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
