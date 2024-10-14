// repository/sqlc/cart_repository_impl.go

package sqlc

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type cartRepositoryImpl struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewCartRepositoryImpl(db *sql.DB, logger *zap.Logger) repository.CartRepository {
	return &cartRepositoryImpl{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// BeginTx starts a new database transaction.
func (r *cartRepositoryImpl) BeginTx(ctx context.Context) (*sql.Tx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction", zap.Error(err))
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	return tx, nil
}

// CreateShoppingCart creates a new shopping cart for a user or guest.
// Only one of userID or guestID should be non-nil to maintain mutual exclusivity.
func (r *cartRepositoryImpl) CreateShoppingCart(ctx context.Context, userID, guestID *uuid.UUID) (*model.ShoppingCart, error) {
	var userNull, guestNull uuid.NullUUID

	if userID != nil && guestID != nil {
		return nil, fmt.Errorf("either userID or guestID should be provided, not both")
	}

	if userID != nil {
		userNull = uuid.NullUUID{UUID: *userID, Valid: true}
	}

	if guestID != nil {
		guestNull = uuid.NullUUID{UUID: *guestID, Valid: true}
	}

	params := database.CreateShoppingCartParams{
		UserID:  userNull,
		GuestID: guestNull,
	}

	log.Println("params", params)

	cart, err := r.Queries.CreateShoppingCart(ctx, params)
	if err != nil {
		r.logger.Error("Failed to create shopping cart", zap.Error(err))
		return nil, fmt.Errorf("create shopping cart: %w", err)
	}

	return convertDBShoppingCartToModel(cart), nil
}

// GetShoppingCartByID retrieves a shopping cart by its ID.
func (r *cartRepositoryImpl) GetShoppingCartByID(ctx context.Context, id uuid.UUID) (*model.ShoppingCart, error) {
	cart, err := r.Queries.GetShoppingCartByID(ctx, id)
	if err != nil {
		r.logger.Error("Failed to get shopping cart by ID", zap.Error(err))
		return nil, fmt.Errorf("get shopping cart by ID: %w", err)
	}
	return convertDBShoppingCartToModel(cart), nil
}

// GetShoppingCartByUserID retrieves a shopping cart by user ID.
func (r *cartRepositoryImpl) GetShoppingCartByUserID(ctx context.Context, userID uuid.UUID) (*model.ShoppingCart, error) {
	cart, err := r.Queries.GetShoppingCartByUserID(ctx, uuid.NullUUID{UUID: userID, Valid: true})
	if err != nil {
		r.logger.Error("Failed to get shopping cart by user ID", zap.Error(err))
		return nil, fmt.Errorf("get shopping cart by user ID: %w", err)
	}
	return convertDBShoppingCartToModel(cart), nil
}

// GetCartByGuestID retrieves a shopping cart by guest ID.
func (r *cartRepositoryImpl) GetCartByGuestID(ctx context.Context, guestID uuid.UUID) (*model.ShoppingCart, error) {
	cart, err := r.Queries.GetCartByGuestID(ctx, uuid.NullUUID{UUID: guestID, Valid: true})
	if err != nil {
		r.logger.Error("Failed to get cart by guest ID", zap.Error(err))
		return nil, fmt.Errorf("get cart by guest ID: %w", err)
	}
	return convertDBShoppingCartToModel(cart), nil
}

// DeleteShoppingCart deletes a shopping cart by its ID.
func (r *cartRepositoryImpl) DeleteShoppingCart(ctx context.Context, id uuid.UUID) error {
	err := r.Queries.DeleteShoppingCart(ctx, id)
	if err != nil {
		r.logger.Error("Failed to delete shopping cart", zap.Error(err))
		return fmt.Errorf("delete shopping cart: %w", err)
	}
	return nil
}

// UpdateCartTotals updates the total_items and total_price of a shopping cart.
func (r *cartRepositoryImpl) UpdateCartTotals(ctx context.Context, id uuid.UUID) error {
	err := r.Queries.UpdateCartTotals(ctx, id)
	if err != nil {
		r.logger.Error("Failed to update cart totals", zap.Error(err))
		return fmt.Errorf("update cart totals: %w", err)
	}
	return nil
}

// CalculateCartTotal calculates the total price of items in the cart.
func (r *cartRepositoryImpl) CalculateCartTotal(ctx context.Context, shoppingCartID uuid.UUID) (float64, error) {
	total, err := r.Queries.CalculateCartTotal(ctx, shoppingCartID)
	if err != nil {
		r.logger.Error("Failed to calculate cart total", zap.Error(err))
		return 0, fmt.Errorf("calculate cart total: %w", err)
	}
	// Assuming total_price is stored as numeric(10,2), convert it to float64
	totalFloat, err := strconv.ParseFloat(total, 64)
	if err != nil {
		r.logger.Error("Failed to parse cart total", zap.Error(err))
		return 0, fmt.Errorf("parse cart total: %w", err)
	}
	return totalFloat, nil
}

// ClearCart removes all items from the cart.
func (r *cartRepositoryImpl) ClearCart(ctx context.Context, shoppingCartID uuid.UUID) error {
	err := r.Queries.ClearCart(ctx, shoppingCartID)
	if err != nil {
		r.logger.Error("Failed to clear cart", zap.Error(err))
		return fmt.Errorf("clear cart: %w", err)
	}
	return nil
}

// GetAllCartItems retrieves all items in the cart.
func (r *cartRepositoryImpl) GetAllCartItems(ctx context.Context, shoppingCartID uuid.UUID) ([]*model.CartItem, error) {
	items, err := r.Queries.GetAllCartItems(ctx, shoppingCartID)
	if err != nil {
		r.logger.Error("Failed to get all cart items", zap.Error(err))
		return nil, fmt.Errorf("get all cart items: %w", err)
	}
	return convertDBCartItemsToModel(items), nil
}

// GetCartItem retrieves a specific cart item.
func (r *cartRepositoryImpl) GetCartItem(ctx context.Context, shoppingCartID, productID uuid.UUID) (*model.CartItem, error) {
	item, err := r.Queries.GetCartItem(ctx, database.GetCartItemParams{
		ShoppingCartID: shoppingCartID,
		ProductID:      productID,
	})
	if err != nil {
		r.logger.Error("Failed to get cart item", zap.Error(err))
		return nil, fmt.Errorf("get cart item: %w", err)
	}
	return convertDBCartItemToModel(item), nil
}

// RemoveCartItem removes a specific item from the cart.
func (r *cartRepositoryImpl) RemoveCartItem(ctx context.Context, shoppingCartID, productID uuid.UUID) error {
	err := r.Queries.RemoveCartItem(ctx, database.RemoveCartItemParams{
		ShoppingCartID: shoppingCartID,
		ProductID:      productID,
	})
	if err != nil {
		r.logger.Error("Failed to remove cart item", zap.Error(err))
		return fmt.Errorf("remove cart item: %w", err)
	}
	return nil
}

// UpdateCartItemQuantity updates the quantity of a specific cart item.
func (r *cartRepositoryImpl) UpdateCartItemQuantity(ctx context.Context, shoppingCartID, productID uuid.UUID, quantity int32) error {
	err := r.Queries.UpdateCartItemQuantity(ctx, database.UpdateCartItemQuantityParams{
		ShoppingCartID: shoppingCartID,
		ProductID:      productID,
		Quantity:       quantity,
	})
	if err != nil {
		r.logger.Error("Failed to update cart item quantity", zap.Error(err))
		return fmt.Errorf("update cart item quantity: %w", err)
	}
	return nil
}

// UpdateCartUserID associates a cart with a user ID and nullifies guest_id.
func (r *cartRepositoryImpl) UpdateCartUserID(ctx context.Context, cartID, userID uuid.UUID) error {
	err := r.Queries.UpdateCartUserID(ctx, database.UpdateCartUserIDParams{
		ID:     cartID,
		UserID: uuid.NullUUID{UUID: userID, Valid: true},
	})
	if err != nil {
		r.logger.Error("Failed to update cart user ID", zap.Error(err))
		return fmt.Errorf("update cart user ID: %w", err)
	}
	return nil
}

// UpdateCartGuestID associates a cart with a guest ID and nullifies user_id.
func (r *cartRepositoryImpl) UpdateCartGuestID(ctx context.Context, cartID, guestID uuid.UUID) error {
	err := r.Queries.UpdateCartGuestID(ctx, database.UpdateCartGuestIDParams{
		ID:      cartID,
		GuestID: uuid.NullUUID{UUID: guestID, Valid: true},
	})
	if err != nil {
		r.logger.Error("Failed to update cart guest ID", zap.Error(err))
		return fmt.Errorf("update cart guest ID: %w", err)
	}
	return nil
}

// UpsertCartItem inserts or updates a cart item.
func (r *cartRepositoryImpl) UpsertCartItem(ctx context.Context, shoppingCartID, productID uuid.UUID, quantity int32, price string) (*model.CartItem, error) {
	item, err := r.Queries.UpsertCartItem(ctx, database.UpsertCartItemParams{
		ShoppingCartID: shoppingCartID,
		ProductID:      productID,
		Quantity:       quantity,
		Price:          price,
	})
	if err != nil {
		r.logger.Error("Failed to upsert cart item", zap.Error(err))
		return nil, fmt.Errorf("upsert cart item: %w", err)
	}
	return convertDBUpsertCartItemToModel(item), nil
}

// Helper functions to convert database models to domain models

func convertDBShoppingCartToModel(dbCart database.ShoppingCart) *model.ShoppingCart {
	totalPrice, err := strconv.ParseFloat(dbCart.TotalPrice, 64)
	if err != nil {
		totalPrice = 0.0
	}

	var userID *uuid.UUID
	var guestID *uuid.UUID

	if dbCart.UserID.Valid {
		userID = &dbCart.UserID.UUID
	}

	if dbCart.GuestID.Valid {
		guestID = &dbCart.GuestID.UUID
	}

	return &model.ShoppingCart{
		ID:         dbCart.ID,
		UserID:     userID,
		GuestID:    guestID,
		TotalItems: dbCart.TotalItems,
		TotalPrice: totalPrice,
	}
}

func convertDBCartItemsToModel(dbItems []database.GetAllCartItemsRow) []*model.CartItem {
	items := make([]*model.CartItem, len(dbItems))
	for i, dbItem := range dbItems {
		items[i] = &model.CartItem{
			ID:          dbItem.ID,
			ProductID:   dbItem.ProductID,
			Name:        dbItem.Name,
			Description: dbItem.Description.String,
			ImageURL:    dbItem.ImageUrl.String,
			Quantity:    dbItem.Quantity,
			Price:       dbItem.Price,
		}
	}
	return items
}

func convertDBCartItemToModel(dbItem database.GetCartItemRow) *model.CartItem {
	return &model.CartItem{
		ID:          dbItem.ID,
		ProductID:   dbItem.ProductID,
		Name:        dbItem.Name,
		Description: dbItem.Description.String,
		ImageURL:    dbItem.ImageUrl.String,
		Quantity:    dbItem.Quantity,
		Price:       dbItem.Price,
	}
}

func convertDBUpsertCartItemToModel(dbItem database.UpsertCartItemRow) *model.CartItem {
	return &model.CartItem{
		ID:          dbItem.ID,
		ProductID:   dbItem.ProductID,
		Name:        dbItem.Name,
		Description: dbItem.Description.String,
		ImageURL:    dbItem.ImageUrl.String,
		Quantity:    dbItem.Quantity,
		Price:       dbItem.Price,
	}
}
