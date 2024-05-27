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

func TestCreateProduct(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewProductRepository(db, logger)

	ctx := context.Background()
	productID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()
	categoryID := uuid.New()
	userID := uuid.New()

	// Expectations for the SQL queries
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO products").WithArgs(
		"Test Product",
		sql.NullString{String: "Product Description", Valid: true},
		"100.00",
		sql.NullInt32{Int32: 10, Valid: true},
		uuid.NullUUID{UUID: categoryID, Valid: true},
		uuid.NullUUID{UUID: userID, Valid: true},
		uuid.NullUUID{UUID: userID, Valid: true},
	).WillReturnRows(sqlmock.NewRows([]string{
		"id", "name", "description", "price", "stock", "category_id", "created_at", "updated_at", "is_active", "created_by", "updated_by",
	}).AddRow(productID, "Test Product", "Product Description", "100.00", 10, categoryID, createdAt, updatedAt, true, userID, userID))
	mock.ExpectCommit()

	product, err := repo.CreateProduct(ctx, database.CreateProductParams{
		Name:        "Test Product",
		Description: sql.NullString{String: "Product Description", Valid: true},
		Price:       "100.00",
		Stock:       sql.NullInt32{Int32: 10, Valid: true},
		CategoryID:  uuid.NullUUID{UUID: categoryID, Valid: true},
		CreatedBy:   uuid.NullUUID{UUID: userID, Valid: true},
		UpdatedBy:   uuid.NullUUID{UUID: userID, Valid: true},
	})
	require.NoError(t, err)
	require.NotEmpty(t, product)
	require.Equal(t, "Test Product", product.Name)
	require.Equal(t, "Product Description", product.Description.String)
	require.Equal(t, "100.00", product.Price)
	require.Equal(t, int32(10), product.Stock.Int32)
	require.Equal(t, categoryID, product.CategoryID.UUID)
	require.Equal(t, userID, product.CreatedBy.UUID)
	require.Equal(t, userID, product.UpdatedBy.UUID)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestGetProductByID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewProductRepository(db, logger)

	ctx := context.Background()
	productID := uuid.New()
	categoryID := uuid.New()
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	// Expectations for the SQL queries
	mock.ExpectQuery("SELECT (.+) FROM products WHERE id = \\$1").
		WithArgs(productID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "price", "stock", "category_id", "created_at", "updated_at", "is_active", "created_by", "updated_by",
		}).AddRow(productID, "Test Product", "Product Description", "100.00", 10, categoryID, createdAt, updatedAt, true, userID, userID))

	product, err := repo.GetProductByID(ctx, productID)
	require.NoError(t, err)
	require.NotEmpty(t, product)
	require.Equal(t, "Test Product", product.Name)
	require.Equal(t, "Product Description", product.Description.String)
	require.Equal(t, "100.00", product.Price)
	require.Equal(t, int32(10), product.Stock.Int32)
	require.Equal(t, categoryID, product.CategoryID.UUID)
	require.Equal(t, userID, product.CreatedBy.UUID)
	require.Equal(t, userID, product.UpdatedBy.UUID)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestListProducts(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewProductRepository(db, logger)

	ctx := context.Background()
	productID := uuid.New()
	categoryID := uuid.New()
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()
	limit := int32(10)
	offset := int32(0)

	// Expectations for the SQL queries
	mock.ExpectQuery("SELECT (.+) FROM products ORDER BY name").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "price", "stock", "category_id", "created_at", "updated_at", "is_active", "created_by", "updated_by",
		}).AddRow(productID, "Test Product", "Product Description", "100.00", 10, categoryID, createdAt, updatedAt, true, userID, userID))

	products, err := repo.ListProducts(ctx, limit, offset)
	require.NoError(t, err)
	require.NotEmpty(t, products)
	require.Len(t, products, 1)

	product := products[0]
	require.Equal(t, "Test Product", product.Name)
	require.Equal(t, "Product Description", product.Description.String)
	require.Equal(t, "100.00", product.Price)
	require.Equal(t, int32(10), product.Stock.Int32)
	require.Equal(t, categoryID, product.CategoryID.UUID)
	require.Equal(t, userID, product.CreatedBy.UUID)
	require.Equal(t, userID, product.UpdatedBy.UUID)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUpdateProduct(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewProductRepository(db, logger)

	ctx := context.Background()
	productID := uuid.New()
	categoryID := uuid.New()
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	// Expectations for the SQL queries
	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE products SET (.+) WHERE id = \\$1 RETURNING (.+)").WithArgs(
		productID,
		"Updated Product",
		sql.NullString{String: "Updated Description", Valid: true},
		"150.00",
		sql.NullInt32{Int32: 20, Valid: true},
		uuid.NullUUID{UUID: categoryID, Valid: true},
		uuid.NullUUID{UUID: userID, Valid: true},
	).WillReturnRows(sqlmock.NewRows([]string{
		"id", "name", "description", "price", "stock", "category_id", "created_at", "updated_at", "is_active", "created_by", "updated_by",
	}).AddRow(productID, "Updated Product", "Updated Description", "150.00", 20, categoryID, createdAt, updatedAt, true, userID, userID))
	mock.ExpectCommit()

	params := database.UpdateProductParams{
		ID:          productID,
		Name:        "Updated Product",
		Description: sql.NullString{String: "Updated Description", Valid: true},
		Price:       "150.00",
		Stock:       sql.NullInt32{Int32: 20, Valid: true},
		CategoryID:  uuid.NullUUID{UUID: categoryID, Valid: true},
		UpdatedBy:   uuid.NullUUID{UUID: userID, Valid: true},
	}

	updatedProduct, err := repo.UpdateProduct(ctx, params)
	require.NoError(t, err)
	require.NotEmpty(t, updatedProduct)
	require.Equal(t, "Updated Product", updatedProduct.Name)
	require.Equal(t, "Updated Description", updatedProduct.Description.String)
	require.Equal(t, "150.00", updatedProduct.Price)
	require.Equal(t, int32(20), updatedProduct.Stock.Int32)
	require.Equal(t, categoryID, updatedProduct.CategoryID.UUID)
	require.Equal(t, userID, updatedProduct.UpdatedBy.UUID)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestSoftDeleteProduct(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewProductRepository(db, logger)

	ctx := context.Background()
	productID := uuid.New()

	// Expectations for the SQL queries
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE products SET is_active = FALSE, updated_at = NOW\(\) WHERE id = \$1`).
		WithArgs(productID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.SoftDeleteProduct(ctx, productID)
	require.NoError(t, err)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestGetProductsByCategoryID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewProductRepository(db, logger)

	ctx := context.Background()
	categoryID := uuid.New()
	productID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	// Expectations for the SQL queries
	mock.ExpectQuery(`SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by FROM products WHERE category_id = \$1 ORDER BY name`).
		WithArgs(uuid.NullUUID{UUID: categoryID, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "price", "stock", "category_id", "created_at", "updated_at", "is_active", "created_by", "updated_by",
		}).AddRow(productID, "Test Product", sql.NullString{String: "Product Description", Valid: true}, "100.00", sql.NullInt32{Int32: 10, Valid: true}, categoryID, createdAt, updatedAt, sql.NullBool{Bool: true, Valid: true}, uuid.New(), uuid.New()))

	products, err := repo.GetProductsByCategoryID(ctx, uuid.NullUUID{UUID: categoryID, Valid: true})
	require.NoError(t, err)
	require.NotEmpty(t, products)
	require.Len(t, products, 1)
	require.Equal(t, "Test Product", products[0].Name)
	require.Equal(t, "Product Description", products[0].Description.String)
	require.Equal(t, "100.00", products[0].Price)
	require.Equal(t, int32(10), products[0].Stock.Int32)
	require.Equal(t, categoryID, products[0].CategoryID.UUID)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestSearchProducts(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, mock := setupTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewProductRepository(db, logger)

	ctx := context.Background()
	productID := uuid.New()
	searchTerm := sql.NullString{String: "Test", Valid: true}
	createdAt := time.Now()
	updatedAt := time.Now()

	// Expectations for the SQL queries
	mock.ExpectQuery(`SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by FROM products WHERE \(name ILIKE '%' \|\| \$1 \|\| '%' OR description ILIKE '%' \|\| \$1 \|\| '%'\) ORDER BY name`).
		WithArgs(searchTerm).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "price", "stock", "category_id", "created_at", "updated_at", "is_active", "created_by", "updated_by",
		}).AddRow(productID, "Test Product", sql.NullString{String: "Product Description", Valid: true}, "100.00", sql.NullInt32{Int32: 10, Valid: true}, uuid.New(), createdAt, updatedAt, sql.NullBool{Bool: true, Valid: true}, uuid.New(), uuid.New()))

	products, err := repo.SearchProducts(ctx, searchTerm)
	require.NoError(t, err)
	require.NotEmpty(t, products)
	require.Len(t, products, 1)
	require.Equal(t, "Test Product", products[0].Name)
	require.Equal(t, "Product Description", products[0].Description.String)
	require.Equal(t, "100.00", products[0].Price)
	require.Equal(t, int32(10), products[0].Stock.Int32)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
