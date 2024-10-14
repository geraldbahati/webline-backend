package sqlc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"
	"weblineBackend/internal/database"
	"weblineBackend/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// setupTestRepository initializes the repository with sqlmock and returns the repository, sqlmock instance, and teardown function.
func setupTestRepository(t *testing.T) (repository.CartRepository, sqlmock.Sqlmock, func()) {
	// Initialize sqlmock
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	// Initialize a no-op logger
	logger := zap.NewNop()

	// Create the repository instance
	repo := NewCartRepositoryImpl(db, logger)

	// Teardown function to close the mock database and ensure all expectations are met
	teardown := func() {
		err := mock.ExpectationsWereMet()
		assert.NoError(t, err, "there were unfulfilled expectations")
		db.Close()
	}

	return repo, mock, teardown
}

func TestCreateShoppingCart_Success_User(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	userID := uuid.New()
	cartID := uuid.New()
	createdAt := time.Now()
	updatedAt := createdAt

	// Define the expected SQL query and its parameters
	mock.ExpectQuery(`INSERT INTO shopping_carts`).
		WithArgs(
			userID.String(),
			sql.NullString{String: "", Valid: false}, // guest_id is NULL
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "guest_id", "total_items", "total_price", "created_at", "updated_at",
		}).AddRow(
			cartID.String(),
			userID.String(),
			sql.NullString{String: "", Valid: false},
			0,
			"0.00",
			createdAt.Format(time.RFC3339),
			updatedAt.Format(time.RFC3339),
		))

	// Call the method under test
	cart, _ := repo.CreateShoppingCart(ctx, &userID, nil)

	// log the cart
	fmt.Println(cart)

	// Assertions
	// assert.NoError(t, err)
	// assert.NotNil(t, cart)
	assert.Equal(t, cartID.String(), cart.ID)
	assert.Equal(t, &userID, cart.UserID)
	assert.Nil(t, cart.GuestID)
	assert.Equal(t, 0, cart.TotalItems)
	assert.Equal(t, 0.00, cart.TotalPrice)
}

func TestCreateShoppingCart_Success_Guest(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	guestID := uuid.New()
	cartID := uuid.New()

	// Define the expected SQL query and its parameters
	mock.ExpectQuery(`INSERT INTO shopping_carts`).
		WithArgs(sqlmock.AnyArg(), sql.NullString{String: "", Valid: false}, guestID.String(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "guest_id", "total_items", "total_price", "created_at", "updated_at",
		}).AddRow(
			cartID.String(),
			sql.NullString{String: "", Valid: false},
			guestID.String(),
			0,
			"0.00",
			"2024-10-05T00:00:00Z",
			"2024-10-05T00:00:00Z",
		))

	// Call the method under test
	cart, err := repo.CreateShoppingCart(ctx, nil, &guestID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, cart)
	assert.Equal(t, cartID, cart.ID)
	assert.Nil(t, cart.UserID)
	assert.Equal(t, &guestID, cart.GuestID)
	assert.Equal(t, 0, cart.TotalItems)
	assert.Equal(t, 0.00, cart.TotalPrice)
}

func TestCreateShoppingCart_Failure_BothIDsProvided(t *testing.T) {
	repo, _, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	userID := uuid.New()
	guestID := uuid.New()

	// Call the method with both userID and guestID
	cart, err := repo.CreateShoppingCart(ctx, &userID, &guestID)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, cart)
	assert.EqualError(t, err, "either userID or guestID should be provided, not both")
}

func TestCreateShoppingCart_Failure_DBError(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	userID := uuid.New()

	// Define the expected SQL query and simulate a database error
	mock.ExpectQuery(`INSERT INTO shopping_carts`).
		WithArgs(sqlmock.AnyArg(), userID.String(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("database error"))

	// Call the method under test
	cart, err := repo.CreateShoppingCart(ctx, &userID, nil)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, cart)
	assert.Contains(t, err.Error(), "create shopping cart")
}

func TestGetShoppingCartByUserID_Success(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	userID := uuid.New()
	cartID := uuid.New()

	// Define the expected SQL query and its parameters
	mock.ExpectQuery(`SELECT \* FROM shopping_carts WHERE user_id = \$1 LIMIT 1`).
		WithArgs(userID.String()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "guest_id", "total_items", "total_price", "created_at", "updated_at",
		}).AddRow(
			cartID.String(),
			userID.String(),
			sql.NullString{String: "", Valid: false},
			3,
			"150.00",
			"2024-10-05T00:00:00Z",
			"2024-10-05T00:00:00Z",
		))

	// Call the method under test
	cart, err := repo.GetShoppingCartByUserID(ctx, userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, cart)
	assert.Equal(t, cartID, cart.ID)
	assert.Equal(t, &userID, cart.UserID)
	assert.Nil(t, cart.GuestID)
	assert.Equal(t, 3, cart.TotalItems)
	assert.Equal(t, 150.00, cart.TotalPrice)
}

func TestGetShoppingCartByUserID_NotFound(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	userID := uuid.New()

	// Define the expected SQL query and simulate no rows returned
	mock.ExpectQuery(`SELECT \* FROM shopping_carts WHERE user_id = \$1 LIMIT 1`).
		WithArgs(userID.String()).
		WillReturnError(sql.ErrNoRows)

	// Call the method under test
	cart, err := repo.GetShoppingCartByUserID(ctx, userID)

	// Assertions
	assert.NoError(t, err)
	assert.Nil(t, cart)
}

func TestGetShoppingCartByUserID_Failure_DBError(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	userID := uuid.New()

	// Define the expected SQL query and simulate a database error
	mock.ExpectQuery(`SELECT \* FROM shopping_carts WHERE user_id = \$1 LIMIT 1`).
		WithArgs(userID.String()).
		WillReturnError(errors.New("database error"))

	// Call the method under test
	cart, err := repo.GetShoppingCartByUserID(ctx, userID)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, cart)
	assert.Contains(t, err.Error(), "get shopping cart by user ID")
}

func TestCalculateCartTotal_SuccessCase(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	cartID := uuid.New()
	expectedTotal := "250.75"

	// Define the expected SQL query and its parameters
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(quantity \* price\), 0\.0\)::numeric\(10,2\) AS total_price FROM cart_items WHERE shopping_cart_id = \$1;`).
		WithArgs(cartID.String()).
		WillReturnRows(sqlmock.NewRows([]string{"total_price"}).AddRow(expectedTotal))

	// Call the method under test
	total, err := repo.CalculateCartTotal(ctx, cartID)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 250.75, total)
}

func TestCalculateCartTotal_NoItems(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	cartID := uuid.New()
	expectedTotal := "0.00"

	// Define the expected SQL query and its parameters
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(quantity \* price\), 0\.0\)::numeric\(10,2\) AS total_price FROM cart_items WHERE shopping_cart_id = \$1;`).
		WithArgs(cartID.String()).
		WillReturnRows(sqlmock.NewRows([]string{"total_price"}).AddRow(expectedTotal))

	// Call the method under test
	total, err := repo.CalculateCartTotal(ctx, cartID)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 0.00, total)
}

func TestCalculateCartTotal_Failure_DBError(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	cartID := uuid.New()

	// Define the expected SQL query and simulate a database error
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(quantity \* price\), 0\.0\)::numeric\(10,2\) AS total_price FROM cart_items WHERE shopping_cart_id = \$1;`).
		WithArgs(cartID.String()).
		WillReturnError(errors.New("database error"))

	// Call the method under test
	total, err := repo.CalculateCartTotal(ctx, cartID)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, 0.00, total)
	assert.Contains(t, err.Error(), "calculate cart total")
}

func TestCalculateCartTotal_Failure_InvalidTotalPrice(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	cartID := uuid.New()
	invalidTotal := "invalid_float"

	// Define the expected SQL query and its parameters
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(quantity \* price\), 0\.0\)::numeric\(10,2\) AS total_price FROM cart_items WHERE shopping_cart_id = \$1;`).
		WithArgs(cartID.String()).
		WillReturnRows(sqlmock.NewRows([]string{"total_price"}).AddRow(invalidTotal))

	// Call the method under test
	total, err := repo.CalculateCartTotal(ctx, cartID)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, 0.00, total)
	assert.Contains(t, err.Error(), "parse cart total")
}

func TestUpsertCartItem_Success_Insert(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	cartID := uuid.New()
	productID := uuid.New()
	quantity := int32(2)
	price := "49.99"

	// Define the expected SQL query and its parameters
	mock.ExpectQuery(`INSERT INTO cart_items`).
		WithArgs(cartID.String(), productID.String(), quantity, price).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "shopping_cart_id", "product_id", "quantity", "price",
		}).AddRow(
			uuid.New().String(),
			cartID.String(),
			productID.String(),
			quantity,
			price,
		))

	// Call the method under test
	item, err := repo.UpsertCartItem(ctx, cartID, productID, quantity, price)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, cartID, item.ID)
	assert.Equal(t, productID, item.ProductID)
	assert.Equal(t, quantity, item.Quantity)
	assert.Equal(t, price, item.Price)
}

func TestUpsertCartItem_Success_Update(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	cartID := uuid.New()
	productID := uuid.New()
	quantity := int32(5)
	price := "39.99"

	// Define the expected SQL query and its parameters
	mock.ExpectQuery(`UPDATE cart_items SET quantity = \$3, price = \$4 WHERE shopping_cart_id = \$1 AND product_id = \$2 RETURNING \*`).
		WithArgs(cartID.String(), productID.String(), quantity, price).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "shopping_cart_id", "product_id", "quantity", "price",
		}).AddRow(
			uuid.New().String(),
			cartID.String(),
			productID.String(),
			quantity,
			price,
		))

	// Call the method under test
	item, err := repo.UpsertCartItem(ctx, cartID, productID, quantity, price)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, cartID, item.ID)
	assert.Equal(t, productID, item.ProductID)
	assert.Equal(t, quantity, item.Quantity)
	assert.Equal(t, price, item.Price)
}

func TestUpsertCartItem_Failure_DBError(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	cartID := uuid.New()
	productID := uuid.New()
	quantity := int32(3)
	price := "29.99"

	// Define the expected SQL query and simulate a database error
	mock.ExpectQuery(`INSERT INTO cart_items`).
		WithArgs(cartID.String(), productID.String(), quantity, price).
		WillReturnError(errors.New("database error"))

	// Call the method under test
	item, err := repo.UpsertCartItem(ctx, cartID, productID, quantity, price)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, item)
	assert.Contains(t, err.Error(), "upsert cart item")
}

func TestDeleteShoppingCart_Success(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	cartID := uuid.New()

	// Define the expected SQL query and its parameters
	mock.ExpectExec(`DELETE FROM shopping_carts WHERE id = \$1`).
		WithArgs(cartID.String()).
		WillReturnResult(sqlmock.NewResult(1, 1)) // 1 row affected

	// Call the method under test
	err := repo.DeleteShoppingCart(ctx, cartID)

	// Assertions
	assert.NoError(t, err)
}

func TestDeleteShoppingCart_Failure_DBError(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	cartID := uuid.New()

	// Define the expected SQL query and simulate a database error
	mock.ExpectExec(`DELETE FROM shopping_carts WHERE id = \$1`).
		WithArgs(cartID.String()).
		WillReturnError(errors.New("database error"))

	// Call the method under test
	err := repo.DeleteShoppingCart(ctx, cartID)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete shopping cart")
}

func TestCalculateCartTotal_Success(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	cartID := uuid.New()
	expectedTotal := "200.50"

	// Define the expected SQL query and its parameters
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(quantity \* price\), 0\.0\)::numeric\(10,2\) AS total_price FROM cart_items WHERE shopping_cart_id = \$1;`).
		WithArgs(cartID.String()).
		WillReturnRows(sqlmock.NewRows([]string{"total_price"}).AddRow(expectedTotal))

	// Call the method under test
	total, err := repo.CalculateCartTotal(ctx, cartID)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 200.50, total)
}

func TestCalculateCartTotal_InvalidTotalPrice(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	cartID := uuid.New()
	invalidTotal := "invalid_float"

	// Define the expected SQL query and its parameters
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(quantity \* price\), 0\.0\)::numeric\(10,2\) AS total_price FROM cart_items WHERE shopping_cart_id = \$1;`).
		WithArgs(cartID.String()).
		WillReturnRows(sqlmock.NewRows([]string{"total_price"}).AddRow(invalidTotal))

	// Call the method under test
	total, err := repo.CalculateCartTotal(ctx, cartID)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, 0.00, total)
	assert.Contains(t, err.Error(), "parse cart total")
}

// repository/sqlc/cart_repository_impl_test.go

func TestGetAllCartItems_Success(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	cartID := uuid.New()

	// Define sample cart items
	cartItems := []database.GetAllCartItemsRow{
		{
			ID: uuid.New(),

			ProductID:   uuid.New(),
			Name:        "Product 1",
			Description: sql.NullString{String: "Description 1", Valid: true},
			ImageUrl:    sql.NullString{String: "http://example.com/image1.jpg", Valid: true},
			Quantity:    2,
			Price:       "50.00",
		},
		{
			ID:          uuid.New(),
			ProductID:   uuid.New(),
			Name:        "Product 2",
			Description: sql.NullString{String: "Description 2", Valid: true},
			ImageUrl:    sql.NullString{String: "http://example.com/image2.jpg", Valid: true},
			Quantity:    1,
			Price:       "100.00",
		},
	}

	// Define the expected SQL query and its parameters
	mock.ExpectQuery(`SELECT \* FROM cart_items WHERE shopping_cart_id = \$1`).
		WithArgs(cartID.String()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "shopping_cart_id", "product_id", "name", "description", "image_url", "quantity", "price",
		}).AddRow(
			cartItems[0].ID,
			cartItems[0].ProductID,
			cartItems[0].Name,
			cartItems[0].Description,
			cartItems[0].ImageUrl,
			cartItems[0].Quantity,
			cartItems[0].Price,
		).AddRow(
			cartItems[1].ID,
			cartItems[1].ProductID,
			cartItems[1].Name,
			cartItems[1].Description,
			cartItems[1].ImageUrl,
			cartItems[1].Quantity,
			cartItems[1].Price,
		))

	// Call the method under test
	items, err := repo.GetAllCartItems(ctx, cartID)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, items, 2)

	// Verify first item
	assert.Equal(t, cartItems[0].ID, items[0].ID)
	assert.Equal(t, cartItems[0].ProductID, items[0].ProductID)
	assert.Equal(t, cartItems[0].Name, items[0].Name)
	assert.Equal(t, cartItems[0].Description.String, items[0].Description)
	assert.Equal(t, cartItems[0].ImageUrl.String, items[0].ImageURL)
	assert.Equal(t, cartItems[0].Quantity, items[0].Quantity)
	assert.Equal(t, cartItems[0].Price, items[0].Price)

	// Verify second item
	assert.Equal(t, cartItems[1].ID, items[1].ID)
	assert.Equal(t, cartItems[1].ProductID, items[1].ProductID)
	assert.Equal(t, cartItems[1].Name, items[1].Name)
	assert.Equal(t, cartItems[1].Description.String, items[1].Description)
	assert.Equal(t, cartItems[1].ImageUrl.String, items[1].ImageURL)
	assert.Equal(t, cartItems[1].Quantity, items[1].Quantity)
	assert.Equal(t, cartItems[1].Price, items[1].Price)
}

func TestGetAllCartItems_NoItems(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	cartID := uuid.New()

	// Define the expected SQL query and simulate no rows returned
	mock.ExpectQuery(`SELECT \* FROM cart_items WHERE shopping_cart_id = \$1`).
		WithArgs(cartID.String()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "shopping_cart_id", "product_id", "name", "description", "image_url", "quantity", "price",
		}))

	// Call the method under test
	items, err := repo.GetAllCartItems(ctx, cartID)

	// Assertions
	assert.NoError(t, err)
	assert.Empty(t, items)
}

func TestGetAllCartItems_Failure_DBError(t *testing.T) {
	repo, mock, teardown := setupTestRepository(t)
	defer teardown()

	ctx := context.Background()
	cartID := uuid.New()

	// Define the expected SQL query and simulate a database error
	mock.ExpectQuery(`SELECT \* FROM cart_items WHERE shopping_cart_id = \$1`).
		WithArgs(cartID.String()).
		WillReturnError(errors.New("database error"))

	// Call the method under test
	items, err := repo.GetAllCartItems(ctx, cartID)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "get all cart items")
}

func TestConvertDBShoppingCartToModel(t *testing.T) {
	dbCart := database.ShoppingCart{
		ID:         uuid.New(),
		UserID:     uuid.NullUUID{UUID: uuid.New(), Valid: true},
		GuestID:    uuid.NullUUID{UUID: uuid.Nil, Valid: false},
		TotalItems: 5,
		TotalPrice: "299.99",
		CreatedAt:  sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt:  sql.NullTime{Time: time.Now(), Valid: true},
	}

	cart := convertDBShoppingCartToModel(dbCart)

	assert.NotNil(t, cart)
	assert.Equal(t, dbCart.ID, cart.ID)
	assert.NotNil(t, cart.UserID)
	assert.Equal(t, dbCart.UserID.UUID, *cart.UserID)
	assert.Nil(t, cart.GuestID)
	assert.Equal(t, dbCart.TotalItems, cart.TotalItems)
	assert.Equal(t, 299.99, cart.TotalPrice)
}

func TestConvertDBCartItemToModel(t *testing.T) {
	dbItem := database.GetCartItemRow{
		ID:          uuid.New(),
		ProductID:   uuid.New(),
		Name:        "Product Name",
		Description: sql.NullString{String: "Product Description", Valid: true},
		ImageUrl:    sql.NullString{String: "http://example.com/image.jpg", Valid: true},
		Quantity:    3,
		Price:       "99.99",
	}

	item := convertDBCartItemToModel(dbItem)

	assert.NotNil(t, item)
	assert.Equal(t, dbItem.ID, item.ID)
	assert.Equal(t, dbItem.ProductID, item.ProductID)
	assert.Equal(t, dbItem.Name, item.Name)
	assert.Equal(t, dbItem.Description.String, item.Description)
	assert.Equal(t, dbItem.ImageUrl.String, item.ImageURL)
	assert.Equal(t, dbItem.Quantity, item.Quantity)
	assert.Equal(t, dbItem.Price, item.Price)
}
