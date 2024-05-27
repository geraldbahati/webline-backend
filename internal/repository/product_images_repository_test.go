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

func TestCreateProductImage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewProductImageRepository(db, logger)

	ctx := context.Background()
	imageID := uuid.New()
	productID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	// Expectations for the SQL queries
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO product_images \\(product_id, image_url\\) VALUES \\(\\$1, \\$2\\) RETURNING id, product_id, image_url, created_at, updated_at").
		WithArgs(productID, "http://example.com/image.jpg").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "product_id", "image_url", "created_at", "updated_at",
		}).AddRow(imageID, productID, "http://example.com/image.jpg", createdAt, updatedAt))
	mock.ExpectCommit()

	productImage, err := repo.CreateProductImage(ctx, database.CreateProductImageParams{
		ProductID: uuid.NullUUID{UUID: productID, Valid: true},
		ImageUrl:  "http://example.com/image.jpg",
	})
	require.NoError(t, err)
	require.NotEmpty(t, productImage)
	require.Equal(t, productID, productImage.ProductID.UUID)
	require.Equal(t, "http://example.com/image.jpg", productImage.ImageUrl)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestGetProductImageByID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewProductImageRepository(db, logger)

	ctx := context.Background()
	imageID := uuid.New()
	productID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	// Expectations for the SQL queries
	mock.ExpectQuery("SELECT id, product_id, image_url, created_at, updated_at FROM product_images WHERE id = \\$1").
		WithArgs(imageID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "product_id", "image_url", "created_at", "updated_at",
		}).AddRow(imageID, productID, "http://example.com/image.jpg", createdAt, updatedAt))

	productImage, err := repo.GetProductImageByID(ctx, imageID)
	require.NoError(t, err)
	require.NotEmpty(t, productImage)
	require.Equal(t, productID, productImage.ProductID.UUID)
	require.Equal(t, "http://example.com/image.jpg", productImage.ImageUrl)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestListProductImagesByProductID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewProductImageRepository(db, logger)

	ctx := context.Background()
	productID := uuid.New()
	imageID1 := uuid.New()
	imageID2 := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	// Expectations for the SQL queries
	mock.ExpectQuery("SELECT id, product_id, image_url, created_at, updated_at FROM product_images WHERE product_id = \\$1").
		WithArgs(uuid.NullUUID{UUID: productID, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "product_id", "image_url", "created_at", "updated_at",
		}).AddRow(imageID1, productID, "http://example.com/image1.jpg", createdAt, updatedAt).
			AddRow(imageID2, productID, "http://example.com/image2.jpg", createdAt, updatedAt))

	productImages, err := repo.ListProductImagesByProductID(ctx, uuid.NullUUID{UUID: productID, Valid: true})
	require.NoError(t, err)
	require.NotEmpty(t, productImages)
	require.Len(t, productImages, 2)
	require.Equal(t, "http://example.com/image1.jpg", productImages[0].ImageUrl)
	require.Equal(t, "http://example.com/image2.jpg", productImages[1].ImageUrl)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUpdateProductImage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewProductImageRepository(db, logger)

	ctx := context.Background()
	imageID := uuid.New()
	productID := uuid.New()
	updatedAt := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE product_images SET image_url = \\$2, updated_at = NOW\\(\\) WHERE id = \\$1 RETURNING id, product_id, image_url, created_at, updated_at").
		WithArgs(imageID, "http://example.com/updated_image.jpg").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "product_id", "image_url", "created_at", "updated_at",
		}).AddRow(imageID, productID, "http://example.com/updated_image.jpg", updatedAt, updatedAt))
	mock.ExpectCommit()

	updatedImage, err := repo.UpdateProductImage(ctx, database.UpdateProductImageParams{
		ID:       imageID,
		ImageUrl: "http://example.com/updated_image.jpg",
	})
	require.NoError(t, err)
	require.NotEmpty(t, updatedImage)
	require.Equal(t, "http://example.com/updated_image.jpg", updatedImage.ImageUrl)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestDeleteProductImage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(db)

	repo := NewProductImageRepository(db, logger)

	ctx := context.Background()
	imageID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM product_images WHERE id = \\$1").
		WithArgs(imageID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.DeleteProductImage(ctx, imageID)
	require.NoError(t, err)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
