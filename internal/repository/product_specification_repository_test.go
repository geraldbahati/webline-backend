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

func TestCreateProductSpecification(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewProductSpecificationRepository(db, logger)

	ctx := context.Background()
	specID := uuid.New()
	productID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO product_specifications \\(product_id, spec_name, spec_value\\) VALUES \\(\\$1, \\$2, \\$3\\) RETURNING id, product_id, spec_name, spec_value, created_at, updated_at").
		WithArgs(productID, "Weight", "1kg").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "product_id", "spec_name", "spec_value", "created_at", "updated_at",
		}).AddRow(specID, productID, "Weight", "1kg", createdAt, updatedAt))
	mock.ExpectCommit()

	specification, err := repo.CreateProductSpecification(ctx, database.CreateProductSpecificationParams{
		ProductID: uuid.NullUUID{UUID: productID, Valid: true},
		SpecName:  "Weight",
		SpecValue: "1kg",
	})
	require.NoError(t, err)
	require.NotEmpty(t, specification)
	require.Equal(t, "Weight", specification.SpecName)
	require.Equal(t, "1kg", specification.SpecValue)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestGetProductSpecificationByID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewProductSpecificationRepository(db, logger)

	ctx := context.Background()
	specID := uuid.New()
	productID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	mock.ExpectQuery("SELECT id, product_id, spec_name, spec_value, created_at, updated_at FROM product_specifications WHERE id = \\$1").
		WithArgs(specID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "product_id", "spec_name", "spec_value", "created_at", "updated_at",
		}).AddRow(specID, productID, "Weight", "1kg", createdAt, updatedAt))

	specification, err := repo.GetProductSpecificationByID(ctx, specID)
	require.NoError(t, err)
	require.NotEmpty(t, specification)
	require.Equal(t, "Weight", specification.SpecName)
	require.Equal(t, "1kg", specification.SpecValue)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestListProductSpecificationsByProductID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewProductSpecificationRepository(db, logger)

	ctx := context.Background()
	productID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	mock.ExpectQuery("SELECT id, product_id, spec_name, spec_value, created_at, updated_at FROM product_specifications WHERE product_id = \\$1").
		WithArgs(productID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "product_id", "spec_name", "spec_value", "created_at", "updated_at",
		}).AddRow(uuid.New(), productID, "Weight", "1kg", createdAt, updatedAt).
			AddRow(uuid.New(), productID, "Dimensions", "10x10x10cm", createdAt, updatedAt))

	specifications, err := repo.ListProductSpecificationsByProductID(ctx, uuid.NullUUID{UUID: productID, Valid: true})
	require.NoError(t, err)
	require.NotEmpty(t, specifications)
	require.Len(t, specifications, 2)
	require.Equal(t, "Weight", specifications[0].SpecName)
	require.Equal(t, "1kg", specifications[0].SpecValue)
	require.Equal(t, "Dimensions", specifications[1].SpecName)
	require.Equal(t, "10x10x10cm", specifications[1].SpecValue)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUpdateProductSpecification(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewProductSpecificationRepository(db, logger)

	ctx := context.Background()
	specID := uuid.New()
	productID := uuid.New()
	updatedAt := time.Now()

	// Expectations for the SQL queries
	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE product_specifications SET spec_name = \\$2, spec_value = \\$3, updated_at = NOW\\(\\) WHERE id = \\$1 RETURNING id, product_id, spec_name, spec_value, created_at, updated_at").
		WithArgs(specID, "New Spec Name", "New Spec Value").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "product_id", "spec_name", "spec_value", "created_at", "updated_at",
		}).AddRow(specID, productID, "New Spec Name", "New Spec Value", time.Now(), updatedAt))
	mock.ExpectCommit()

	updatedSpec, err := repo.UpdateProductSpecification(ctx, database.UpdateProductSpecificationParams{
		ID:        specID,
		SpecName:  "New Spec Name",
		SpecValue: "New Spec Value",
	})
	require.NoError(t, err)
	require.NotEmpty(t, updatedSpec)
	require.Equal(t, "New Spec Name", updatedSpec.SpecName)
	require.Equal(t, "New Spec Value", updatedSpec.SpecValue)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestDeleteProductSpecification(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewProductSpecificationRepository(db, logger)

	ctx := context.Background()
	specID := uuid.New()

	// Expectations for the SQL queries
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM product_specifications WHERE id = \\$1").
		WithArgs(specID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.DeleteProductSpecification(ctx, specID)
	require.NoError(t, err)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
