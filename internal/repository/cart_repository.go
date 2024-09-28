package repository

import (
	"context"
	"database/sql"
	"weblineBackend/internal/model"

	"github.com/google/uuid"
)

type CartRepository interface {
	CreateShoppingCart(ctx context.Context, userID, sessionReference *uuid.UUID) (*model.ShoppingCart, error)
	GetShoppingCartByUserID(ctx context.Context, userID uuid.UUID) (*model.ShoppingCart, error)
	GetCartBySessionReference(ctx context.Context, sessionReference uuid.UUID) (*model.ShoppingCart, error)
	DeleteShoppingCart(ctx context.Context, id uuid.UUID) error
	UpdateCartTotals(ctx context.Context, id uuid.UUID) error

	CalculateCartTotal(ctx context.Context, shoppingCartID uuid.UUID) (int64, error)
	ClearCart(ctx context.Context, shoppingCartID uuid.UUID) error
	GetAllCartItems(ctx context.Context, shoppingCartID uuid.UUID) ([]model.CartItem, error)
	GetCartItem(ctx context.Context, shoppingCartID, productID uuid.UUID) (*model.CartItem, error)
	RemoveCartItem(ctx context.Context, shoppingCartID, productID uuid.UUID) error
	UpdateCartItemQuantity(ctx context.Context, shoppingCartID, productID uuid.UUID, quantity int32) error
	UpdateCartUserID(ctx context.Context, cartID, userID uuid.UUID) error
	UpsertCartItem(ctx context.Context, shoppingCartID, productID uuid.UUID, quantity int32, price string) (*model.CartItem, error)
	BeginTx(ctx context.Context) (*sql.Tx, error)
}
