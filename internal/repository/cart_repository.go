// repository/cart_repository.go

package repository

import (
	"context"
	"database/sql"
	"weblineBackend/internal/model"

	"github.com/google/uuid"
)

type CartRepository interface {
	// Shopping Cart Operations
	CreateShoppingCart(ctx context.Context, params CreateShoppingCartParams) (*model.ShoppingCart, error)
	GetShoppingCartByID(ctx context.Context, id uuid.UUID) (*model.ShoppingCart, error)
	GetShoppingCartByUserID(ctx context.Context, userID uuid.UUID) (*model.ShoppingCart, error)
	GetCartByGuestID(ctx context.Context, guestID uuid.UUID) (*model.ShoppingCart, error)
	DeleteShoppingCart(ctx context.Context, id uuid.UUID) error
	UpdateCartTotals(ctx context.Context, id uuid.UUID) error

	// Cart Items Operations
	CalculateCartTotal(ctx context.Context, shoppingCartID uuid.UUID) (float64, error)
	ClearCart(ctx context.Context, shoppingCartID uuid.UUID) error
	GetAllCartItems(ctx context.Context, shoppingCartID uuid.UUID) ([]*model.CartItem, error)
	GetCartItem(ctx context.Context, shoppingCartID, productID uuid.UUID) (*model.CartItem, error)
	RemoveCartItem(ctx context.Context, shoppingCartID, productID uuid.UUID) error
	UpdateCartItemQuantity(ctx context.Context, shoppingCartID, productID uuid.UUID, quantity int32) error
	UpdateCartUserID(ctx context.Context, cartID, userID uuid.UUID) error
	UpdateCartGuestID(ctx context.Context, cartID, guestID uuid.UUID) error
	UpsertCartItem(ctx context.Context, shoppingCartID, productID uuid.UUID, quantity int32, price string) (*model.CartItem, error)
	BeginTx(ctx context.Context) (*sql.Tx, error)

	// Add this new method
	GetCartByOwnerID(ctx context.Context, ownerID uuid.UUID) (*model.ShoppingCart, error)
}

// Add this struct for better parameter handling
type CreateShoppingCartParams struct {
	UserID  *uuid.UUID
	GuestID *uuid.UUID
}
