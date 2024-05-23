package test

import (
	"context"
	"database/sql"
	"go.uber.org/zap"
	"testing"
	"time"
	"weblineBackend/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"weblineBackend/internal/database"
)

func TestCreateUser(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := repository.NewUserRepository(db, logger)

	ctx := context.Background()
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	// Expectations for the SQL queries
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO users").WithArgs(
		"testuser",
		"test@example.com",
		"hashedpassword",
		"Test",
		"User",
		"1234567890",
		"",
		sqlmock.AnyArg(), // DateOfBirth
	).WillReturnRows(sqlmock.NewRows([]string{
		"id", "username", "email", "first_name", "last_name", "phone_number", "profile_image_url", "date_of_birth", "is_active", "created_at", "updated_at", "last_login",
	}).AddRow(userID, "testuser", "test@example.com", "Test", "User", "1234567890", "", sql.NullTime{}, true, createdAt, updatedAt, sql.NullTime{}))
	mock.ExpectCommit()

	user, err := repo.CreateUser(ctx, database.CreateUserParams{
		Username:        "testuser",
		Email:           "test@example.com",
		HashedPassword:  "hashedpassword",
		FirstName:       sql.NullString{String: "Test", Valid: true},
		LastName:        sql.NullString{String: "User", Valid: true},
		PhoneNumber:     sql.NullString{String: "1234567890", Valid: true},
		ProfileImageUrl: sql.NullString{String: "", Valid: true},
		DateOfBirth:     sql.NullTime{},
	})
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, "testuser", user.Username)
	require.Equal(t, "test@example.com", user.Email)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestGetUserByID(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := repository.NewUserRepository(db, logger)

	ctx := context.Background()
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	mock.ExpectQuery("SELECT (.+) FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "first_name", "last_name", "phone_number", "profile_image_url", "date_of_birth", "is_active", "created_at", "updated_at", "last_login",
		}).AddRow(
			userID, "testuser", "test@example.com", "Test", "User", "1234567890", "", sql.NullTime{}, true, createdAt, updatedAt, sql.NullTime{},
		))

	user, err := repo.GetUserByID(ctx, userID)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, "testuser", user.Username)
	require.Equal(t, "test@example.com", user.Email)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestGetUserByUsername(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := repository.NewUserRepository(db, logger)

	ctx := context.Background()
	username := "testuser"
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	mock.ExpectQuery("SELECT (.+) FROM users WHERE username = \\$1").
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "first_name", "last_name", "phone_number", "profile_image_url", "date_of_birth", "is_active", "created_at", "updated_at", "last_login",
		}).AddRow(
			userID, "testuser", "test@example.com", "Test", "User", "1234567890", "", sql.NullTime{}, true, createdAt, updatedAt, sql.NullTime{},
		))

	user, err := repo.GetUserByUsername(ctx, username)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, "testuser", user.Username)
	require.Equal(t, "test@example.com", user.Email)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestGetUserByEmail(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := repository.NewUserRepository(db, logger)

	ctx := context.Background()
	email := "test@example.com"
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	mock.ExpectQuery("SELECT (.+) FROM users WHERE email = \\$1").
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "first_name", "last_name", "phone_number", "profile_image_url", "date_of_birth", "is_active", "created_at", "updated_at", "last_login",
		}).AddRow(
			userID, "testuser", "test@example.com", "Test", "User", "1234567890", "", sql.NullTime{}, true, createdAt, updatedAt, sql.NullTime{},
		))

	user, err := repo.GetUserByEmail(ctx, email)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, "testuser", user.Username)
	require.Equal(t, "test@example.com", user.Email)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUpdateUserProfile(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := repository.NewUserRepository(db, logger)

	ctx := context.Background()
	userID := uuid.New()
	updatedAt := time.Now()

	// Expectations for the SQL queries
	mock.ExpectQuery(`UPDATE users SET first_name = \$2, last_name = \$3, phone_number = \$4, profile_image_url = \$5, date_of_birth = \$6, updated_at = NOW\(\) WHERE id = \$1 RETURNING id, username, email, first_name, last_name, phone_number, profile_image_url, date_of_birth, is_active, created_at, updated_at, last_login`).WithArgs(
		userID, "NewFirstName", "NewLastName", "1234567890", "newprofileimageurl", sql.NullTime{},
	).WillReturnRows(sqlmock.NewRows([]string{
		"id", "username", "email", "first_name", "last_name", "phone_number", "profile_image_url", "date_of_birth", "is_active", "created_at", "updated_at", "last_login",
	}).AddRow(userID, "testuser", "test@example.com", "NewFirstName", "NewLastName", "1234567890", "newprofileimageurl", sql.NullTime{}, true, updatedAt, updatedAt, sql.NullTime{}))

	user, err := repo.UpdateUserProfile(ctx, database.UpdateUserProfileParams{
		ID:              userID,
		FirstName:       sql.NullString{String: "NewFirstName", Valid: true},
		LastName:        sql.NullString{String: "NewLastName", Valid: true},
		PhoneNumber:     sql.NullString{String: "1234567890", Valid: true},
		ProfileImageUrl: sql.NullString{String: "newprofileimageurl", Valid: true},
		DateOfBirth:     sql.NullTime{},
	})
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, sql.NullString{String: "NewFirstName", Valid: true}, user.FirstName)
	require.Equal(t, sql.NullString{String: "NewLastName", Valid: true}, user.LastName)
	require.Equal(t, sql.NullString{String: "1234567890", Valid: true}, user.PhoneNumber)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUpdateUserPassword(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := repository.NewUserRepository(db, logger)

	ctx := context.Background()
	userID := uuid.New()
	updatedAt := time.Now()

	// Expect the UPDATE query with RETURNING clause and allow for comments
	mock.ExpectQuery("(?s).*UPDATE users SET hashed_password = \\$2, updated_at = NOW\\(\\) WHERE id = \\$1 RETURNING id, username, email, first_name, last_name, phone_number, profile_image_url, date_of_birth, is_active, created_at, updated_at, last_login.*").
		WithArgs(userID, "newhashedpassword").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "first_name", "last_name", "phone_number", "profile_image_url", "date_of_birth", "is_active", "created_at", "updated_at", "last_login",
		}).AddRow(
			userID, "testuser", "test@example.com", "Test", "User", "1234567890", "", sql.NullTime{}, true, updatedAt, updatedAt, sql.NullTime{},
		))

	user, err := repo.UpdateUserPassword(ctx, database.UpdateUserPasswordParams{
		ID:             userID,
		HashedPassword: "newhashedpassword",
	})
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, userID, user.ID)
	require.Equal(t, "test@example.com", user.Email)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUpdateUserLastLogin(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := repository.NewUserRepository(db, logger)

	ctx := context.Background()
	userID := uuid.New()
	updatedAt := time.Now()

	// Expect the UPDATE query with RETURNING clause and allow for comments
	mock.ExpectQuery("(?s).*UPDATE users SET last_login = NOW\\(\\) WHERE id = \\$1 RETURNING id, username, email, first_name, last_name, phone_number, profile_image_url, date_of_birth, is_active, created_at, updated_at, last_login.*").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "first_name", "last_name", "phone_number", "profile_image_url", "date_of_birth", "is_active", "created_at", "updated_at", "last_login",
		}).AddRow(
			userID, "testuser", "test@example.com", "Test", "User", "1234567890", "", sql.NullTime{}, true, updatedAt, updatedAt, time.Now(),
		))

	user, err := repo.UpdateUserLastLogin(ctx, userID)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.NotZero(t, user.LastLogin)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestDeactivateUser(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := repository.NewUserRepository(db, logger)

	ctx := context.Background()
	userID := uuid.New()
	updatedAt := time.Now()

	mock.ExpectQuery("(?s).*UPDATE users SET is_active = FALSE, updated_at = NOW\\(\\) WHERE id = \\$1 RETURNING id, username, email, first_name, last_name, phone_number, profile_image_url, date_of_birth, is_active, created_at, updated_at, last_login").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "first_name", "last_name", "phone_number", "profile_image_url", "date_of_birth", "is_active", "created_at", "updated_at", "last_login",
		}).AddRow(
			userID, "testuser", "test@example.com", "Test", "User", "1234567890", "", sql.NullTime{}, false, updatedAt, updatedAt, sql.NullTime{},
		))

	user, err := repo.DeactivateUser(ctx, userID)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.False(t, user.IsActive)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestDeleteUser(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := repository.NewUserRepository(db, logger)

	ctx := context.Background()
	userID := uuid.New()

	mock.ExpectExec("DELETE FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.DeleteUser(ctx, userID)
	require.NoError(t, err)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestListUsers(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := repository.NewUserRepository(db, logger)

	ctx := context.Background()
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	mock.ExpectQuery("SELECT (.+) FROM users ORDER BY created_at DESC LIMIT \\$1 OFFSET \\$2").
		WithArgs(int32(10), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "first_name", "last_name", "phone_number", "profile_image_url", "date_of_birth", "is_active", "created_at", "updated_at", "last_login",
		}).AddRow(
			userID, "testuser", "test@example.com", "Test", "User", "1234567890", "", sql.NullTime{}, true, createdAt, updatedAt, sql.NullTime{},
		))

	users, err := repo.ListUsers(ctx, 10, 0)
	require.NoError(t, err)
	require.NotEmpty(t, users)
	require.Len(t, users, 1)
	require.Equal(t, "testuser", users[0].Username)
	require.Equal(t, "test@example.com", users[0].Email)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
