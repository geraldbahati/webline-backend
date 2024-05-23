package test

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
	"weblineBackend/internal/repository"
)

func TestCreateProductVariant(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := repository.NewProductVariantRepository(db, logger)

	ctx := context.Background()
	variantID := uuid.New()
	productID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	// Expectations for the SQL queries
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO product_variants").WithArgs(
		uuid.NullUUID{UUID: productID, Valid: true},
		"Color",
		"Red",
		sql.NullString{String: "10.00", Valid: true},
	).WillReturnRows(sqlmock.NewRows([]string{
		"id", "product_id", "variant_name", "variant_value", "additional_price", "created_at", "updated_at",
	}).AddRow(variantID, productID, "Color", "Red", "10.00", createdAt, updatedAt))
	mock.ExpectCommit()

	variant, err := repo.CreateProductVariant(ctx, database.CreateProductVariantParams{
		ProductID:       uuid.NullUUID{UUID: productID, Valid: true},
		VariantName:     "Color",
		VariantValue:    "Red",
		AdditionalPrice: sql.NullString{String: "10.00", Valid: true},
	})
	require.NoError(t, err)
	require.NotEmpty(t, variant)
	require.Equal(t, "Color", variant.VariantName)
	require.Equal(t, "Red", variant.VariantValue)
	require.Equal(t, "10.00", variant.AdditionalPrice.String)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestGetProductVariantByID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := repository.NewProductVariantRepository(db, logger)

	ctx := context.Background()
	variantID := uuid.New()
	productID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	// Expectations for the SQL queries
	mock.ExpectQuery("SELECT (.+) FROM product_variants WHERE id = \\$1").
		WithArgs(variantID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "product_id", "variant_name", "variant_value", "additional_price", "created_at", "updated_at",
		}).AddRow(variantID, productID, "Color", "Red", "10.00", createdAt, updatedAt))

	variant, err := repo.GetProductVariantByID(ctx, variantID)
	require.NoError(t, err)
	require.NotEmpty(t, variant)
	require.Equal(t, "Color", variant.VariantName)
	require.Equal(t, "Red", variant.VariantValue)
	require.Equal(t, "10.00", variant.AdditionalPrice.String)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestListProductVariantsByProductID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := repository.NewProductVariantRepository(db, logger)

	ctx := context.Background()
	productID := uuid.New()
	nullProductID := uuid.NullUUID{UUID: productID, Valid: true}
	createdAt := time.Now()
	updatedAt := time.Now()

	// Expectations for the SQL queries
	mock.ExpectQuery("SELECT (.+) FROM product_variants WHERE product_id = \\$1 ORDER BY variant_name").
		WithArgs(nullProductID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "product_id", "variant_name", "variant_value", "additional_price", "created_at", "updated_at",
		}).AddRow(uuid.New(), productID, "Color", "Red", "10.00", createdAt, updatedAt).
			AddRow(uuid.New(), productID, "Size", "L", "15.00", createdAt, updatedAt))

	variants, err := repo.ListProductVariantsByProductID(ctx, nullProductID)
	require.NoError(t, err)
	require.NotEmpty(t, variants)
	require.Len(t, variants, 2)
	require.Equal(t, "Color", variants[0].VariantName)
	require.Equal(t, "Red", variants[0].VariantValue)
	require.Equal(t, "Size", variants[1].VariantName)
	require.Equal(t, "L", variants[1].VariantValue)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUpdateProductVariant(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := repository.NewProductVariantRepository(db, logger)

	ctx := context.Background()
	variantID := uuid.New()
	productID := uuid.New()
	updatedAt := time.Now()

	// Expectations for the SQL queries
	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE product_variants SET variant_name = \\$2, variant_value = \\$3, additional_price = \\$4, updated_at = NOW\\(\\) WHERE id = \\$1 RETURNING id, product_id, variant_name, variant_value, additional_price, created_at, updated_at").
		WithArgs(variantID, "Color", "Blue", "20.00").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "product_id", "variant_name", "variant_value", "additional_price", "created_at", "updated_at",
		}).AddRow(variantID, productID, "Color", "Blue", "20.00", time.Now(), updatedAt))
	mock.ExpectCommit()

	params := database.UpdateProductVariantParams{
		ID:              variantID,
		VariantName:     "Color",
		VariantValue:    "Blue",
		AdditionalPrice: sql.NullString{String: "20.00", Valid: true},
	}

	variant, err := repo.UpdateProductVariant(ctx, params)
	require.NoError(t, err)
	require.NotEmpty(t, variant)
	require.Equal(t, "Color", variant.VariantName)
	require.Equal(t, "Blue", variant.VariantValue)
	require.Equal(t, "20.00", variant.AdditionalPrice.String)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestDeleteProductVariant(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := repository.NewProductVariantRepository(db, logger)

	ctx := context.Background()
	variantID := uuid.New()

	// Expectations for the SQL queries
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM product_variants WHERE id = \\$1").
		WithArgs(variantID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.DeleteProductVariant(ctx, variantID)
	require.NoError(t, err)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
