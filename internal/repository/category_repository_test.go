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

func TestCreateCategory(t *testing.T) {
	// Initialize the logger
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewCategoryRepository(db, logger)

	ctx := context.Background()
	categoryID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	// Expectations for the SQL queries
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO categories").
		WithArgs("Test Category", nil).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "parent_id", "created_at", "updated_at", "is_active"}).
			AddRow(categoryID, "Test Category", nil, createdAt, updatedAt, true))
	mock.ExpectCommit()

	newCategory := database.CreateCategoryParams{
		Name: "Test Category",
	}

	createdCategory, err := repo.CreateCategory(ctx, newCategory)
	require.NoError(t, err)
	require.NotEmpty(t, createdCategory)
	require.Equal(t, "Test Category", createdCategory.Name)
	require.Equal(t, categoryID, createdCategory.ID)
	require.True(t, createdCategory.CreatedAt.Valid)
	require.True(t, createdCategory.UpdatedAt.Valid)
	require.WithinDuration(t, createdAt, createdCategory.CreatedAt.Time, time.Second)
	require.WithinDuration(t, updatedAt, createdCategory.UpdatedAt.Time, time.Second)
	require.Equal(t, true, createdCategory.IsActive)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestCreateCategoryRollback(t *testing.T) {
	// Initialize the logger
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewCategoryRepository(db, logger)

	ctx := context.Background()

	// Expectations for the SQL queries
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO categories").
		WithArgs("Test Category", nil).
		WillReturnError(sql.ErrConnDone) // Simulate an error
	mock.ExpectRollback()

	newCategory := database.CreateCategoryParams{
		Name: "Test Category",
	}

	createdCategory, err := repo.CreateCategory(ctx, newCategory)
	require.Error(t, err)
	require.Empty(t, createdCategory)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestGetCategoryByID(t *testing.T) {
	// Initialize the logger
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewCategoryRepository(db, logger)

	ctx := context.Background()
	categoryID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	// Expectations for the SQL query
	mock.ExpectQuery("SELECT (.+) FROM categories WHERE id = \\$1").
		WithArgs(categoryID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "parent_id", "created_at", "updated_at", "is_active"}).
			AddRow(categoryID, "Test Category", nil, createdAt, updatedAt, true))

	// Call the GetCategoryByID function
	category, err := repo.GetCategoryByID(ctx, categoryID)
	require.NoError(t, err)
	require.NotEmpty(t, category)
	require.Equal(t, "Test Category", category.Name)
	require.Equal(t, categoryID, category.ID)
	require.True(t, category.CreatedAt.Valid)
	require.True(t, category.UpdatedAt.Valid)
	require.WithinDuration(t, createdAt, category.CreatedAt.Time, time.Second)
	require.WithinDuration(t, updatedAt, category.UpdatedAt.Time, time.Second)
	require.Equal(t, true, category.IsActive)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestGetCategories(t *testing.T) {
	// Initialize the logger
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewCategoryRepository(db, logger)

	ctx := context.Background()
	categoryID1 := uuid.New()
	categoryID2 := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	// Expectations for the SQL query
	mock.ExpectQuery("SELECT (.+) FROM categories ORDER BY name").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "parent_id", "created_at", "updated_at", "is_active"}).
			AddRow(categoryID1, "Category 1", nil, createdAt, updatedAt, true).
			AddRow(categoryID2, "Category 2", nil, createdAt, updatedAt, true))

	// Call the GetCategories function
	categories, err := repo.GetCategories(ctx)
	require.NoError(t, err)
	require.Len(t, categories, 2)

	require.Equal(t, "Category 1", categories[0].Name)
	require.Equal(t, categoryID1, categories[0].ID)
	require.True(t, categories[0].CreatedAt.Valid)
	require.True(t, categories[0].UpdatedAt.Valid)
	require.WithinDuration(t, createdAt, categories[0].CreatedAt.Time, time.Second)
	require.WithinDuration(t, updatedAt, categories[0].UpdatedAt.Time, time.Second)
	require.Equal(t, true, categories[0].IsActive)

	require.Equal(t, "Category 2", categories[1].Name)
	require.Equal(t, categoryID2, categories[1].ID)
	require.True(t, categories[1].CreatedAt.Valid)
	require.True(t, categories[1].UpdatedAt.Valid)
	require.WithinDuration(t, createdAt, categories[1].CreatedAt.Time, time.Second)
	require.WithinDuration(t, updatedAt, categories[1].UpdatedAt.Time, time.Second)
	require.Equal(t, true, categories[1].IsActive)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUpdateCategory(t *testing.T) {
	// Initialize the logger
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewCategoryRepository(db, logger)

	ctx := context.Background()
	categoryID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	// Expectations for the SQL query
	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE categories SET name = \\$2, parent_id = \\$3, updated_at = NOW\\(\\) WHERE id = \\$1 RETURNING id, name, parent_id, created_at, updated_at, is_active").
		WithArgs(categoryID, "Updated Category", uuid.NullUUID{}).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "parent_id", "created_at", "updated_at", "is_active"}).
			AddRow(categoryID, "Updated Category", nil, createdAt, updatedAt, true))
	mock.ExpectCommit()

	// Call the UpdateCategory function
	updatedCategory, err := repo.UpdateCategory(ctx, categoryID, "Updated Category", uuid.NullUUID{})
	require.NoError(t, err)
	require.NotEmpty(t, updatedCategory)
	require.Equal(t, "Updated Category", updatedCategory.Name)
	require.Equal(t, categoryID, updatedCategory.ID)
	require.True(t, updatedCategory.CreatedAt.Valid)
	require.True(t, updatedCategory.UpdatedAt.Valid)
	require.WithinDuration(t, createdAt, updatedCategory.CreatedAt.Time, time.Second)
	require.WithinDuration(t, updatedAt, updatedCategory.UpdatedAt.Time, time.Second)
	require.Equal(t, true, updatedCategory.IsActive)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestSoftDeleteCategory(t *testing.T) {
	// Initialize the logger
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewCategoryRepository(db, logger)

	ctx := context.Background()
	categoryID := uuid.New()

	// Expectations for the SQL query
	mock.ExpectExec("UPDATE categories SET is_active = FALSE, updated_at = NOW\\(\\) WHERE id = \\$1").
		WithArgs(categoryID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the SoftDeleteCategory function
	err := repo.SoftDeleteCategory(ctx, categoryID)
	require.NoError(t, err)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestSoftDeleteCategoryFailure(t *testing.T) {
	// Initialize the logger
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewCategoryRepository(db, logger)

	ctx := context.Background()
	categoryID := uuid.New()

	// Expectations for the SQL query
	mock.ExpectExec("UPDATE categories SET is_active = FALSE, updated_at = NOW\\(\\) WHERE id = \\$1").
		WithArgs(categoryID).
		WillReturnError(sql.ErrConnDone) // Simulate an error

	// Call the SoftDeleteCategory function
	err := repo.SoftDeleteCategory(ctx, categoryID)
	require.Error(t, err)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestGetCategoriesByParentID(t *testing.T) {
	// Initialize the logger
	logger, _ := zap.NewDevelopment()

	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewCategoryRepository(db, logger)

	ctx := context.Background()
	parentID := uuid.New()
	categoryID1 := uuid.New()
	categoryID2 := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	// Expectations for the SQL query
	mock.ExpectQuery("SELECT (.+) FROM categories WHERE parent_id = \\$1 ORDER BY name").
		WithArgs(parentID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "parent_id", "created_at", "updated_at", "is_active"}).
			AddRow(categoryID1, "Category 1", parentID, createdAt, updatedAt, true).
			AddRow(categoryID2, "Category 2", parentID, createdAt, updatedAt, true))

	// Call the GetCategoriesByParentID function
	categories, err := repo.GetCategoriesByParentID(ctx, uuid.NullUUID{UUID: parentID, Valid: true})
	require.NoError(t, err)
	require.Len(t, categories, 2)

	require.Equal(t, "Category 1", categories[0].Name)
	require.Equal(t, categoryID1, categories[0].ID)
	require.True(t, categories[0].CreatedAt.Valid)
	require.True(t, categories[0].UpdatedAt.Valid)
	require.WithinDuration(t, createdAt, categories[0].CreatedAt.Time, time.Second)
	require.WithinDuration(t, updatedAt, categories[0].UpdatedAt.Time, time.Second)
	require.Equal(t, true, categories[0].IsActive)

	require.Equal(t, "Category 2", categories[1].Name)
	require.Equal(t, categoryID2, categories[1].ID)
	require.True(t, categories[1].CreatedAt.Valid)
	require.True(t, categories[1].UpdatedAt.Valid)
	require.WithinDuration(t, createdAt, categories[1].CreatedAt.Time, time.Second)
	require.WithinDuration(t, updatedAt, categories[1].UpdatedAt.Time, time.Second)
	require.Equal(t, true, categories[1].IsActive)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
