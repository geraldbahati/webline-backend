package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CartRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewCartRepository(db *sql.DB, logger *zap.Logger) *CartRepository {
	return &CartRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided functiopn
func (r *CartRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	q := database.New(tx)
	if err := fn(q); err != nil {
		r.logger.Error("transaction failed, rolling back", zap.Error(err))
		if rbErr := tx.Rollback(); rbErr != nil {
			r.logger.Error("rollback failed", zap.Error(rbErr))
			return fmt.Errorf("rollback transaction: %w", rbErr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// AddToCart adds an item to the cart or updates the quantity if it already exists
func (r *CartRepository) AddToCart(ctx context.Context, cartID, productID uuid.NullUUID, quantity int32, price float64) error {
	return r.execTx(ctx, func(q *database.Queries) error {
		// Check if the item exists in the cart
		_, err := q.GetCartItem(ctx, database.GetCartItemParams{
			ShoppingCartID: cartID,
			ProductID:      productID,
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("get cart item: %w", err)
		}

		// Insert or update the cart item
		_, err = q.UpsertCartItem(ctx, database.UpsertCartItemParams{
			ShoppingCartID: cartID,
			ProductID:      productID,
			Quantity:       quantity,
			Price:          strconv.FormatFloat(price, 'f', -1, 64),
		})
		if err != nil {
			return fmt.Errorf("upsert cart item: %w", err)
		}

		r.logger.Info("added/updated item in cart", zap.String("cartID", cartID.UUID.String()), zap.String("productID", productID.UUID.String()), zap.Int("quantity", int(quantity)), zap.Float64("price", price))
		return nil
	})
}

// RemoveFromCart removes an item from the cart
func (r *CartRepository) RemoveFromCart(ctx context.Context, cartID, productID uuid.NullUUID) error {
	return r.execTx(ctx, func(q *database.Queries) error {
		err := q.RemoveCartItem(ctx, database.RemoveCartItemParams{
			ShoppingCartID: cartID,
			ProductID:      productID,
		})
		if err != nil {
			return fmt.Errorf("delete cart item: %w", err)
		}

		r.logger.Info("removed item from cart", zap.String("cartID", cartID.UUID.String()), zap.String("productID", productID.UUID.String()))
		return nil
	})
}

// GetCartItems retrieves all items in the cart
func (r *CartRepository) GetCartItems(ctx context.Context, cartID uuid.NullUUID) ([]database.GetAllCartItemsRow, error) {
	items, err := r.GetAllCartItems(ctx, cartID)
	if err != nil {
		return nil, fmt.Errorf("get cart items: %w", err)
	}

	return items, nil
}

// UpdateCartItemQuantity updates the quantity of an item in the cart
func (r *CartRepository) UpdateCartItemQuantity(ctx context.Context, cartID, productID uuid.NullUUID, quantity int32) error {
	return r.execTx(ctx, func(q *database.Queries) error {
		err := q.UpdateCartItemQuantity(ctx, database.UpdateCartItemQuantityParams{
			ShoppingCartID: cartID,
			ProductID:      productID,
			Quantity:       quantity,
		})
		if err != nil {
			return fmt.Errorf("update cart item quantity: %w", err)
		}

		r.logger.Info("updated item quantity in cart", zap.String("cartID", cartID.UUID.String()), zap.String("productID", productID.UUID.String()), zap.Int("quantity", int(quantity)))
		return nil
	})
}

// ClearCart removes all items from the cart
func (r *CartRepository) ClearCart(ctx context.Context, cartID uuid.NullUUID) error {
	return r.execTx(ctx, func(q *database.Queries) error {
		err := q.ClearCart(ctx, cartID)
		if err != nil {
			return fmt.Errorf("clear cart: %w", err)
		}

		r.logger.Info("cleared cart", zap.String("cartID", cartID.UUID.String()))
		return nil
	})
}

// CalculateCartTotal calculates the total price and number of items in the cart
func (r *CartRepository) CalculateCartTotal(ctx context.Context, cartID uuid.NullUUID) (float64, error) {
	// get all items in the cart
	items, err := r.GetAllCartItems(ctx, cartID)
	if err != nil {
		return 0, fmt.Errorf("get cart items: %w", err)
	}

	// calculate the total price
	var total float64
	for _, item := range items {
		price, err := strconv.ParseFloat(item.Price, 64)
		if err != nil {
			return 0, fmt.Errorf("parse price: %w", err)
		}
		total += price * float64(item.Quantity)
	}

	return total, nil
}

// CreateShoppingCart creates a new shopping cart
func (r *CartRepository) CreateShoppingCart(ctx context.Context, userID uuid.NullUUID) (model.ShoppingCart, error) {
	cartID, err := r.Queries.CreateShoppingCart(ctx, userID)
	if err != nil {
		return model.ShoppingCart{}, fmt.Errorf("create shopping cart: %w", err)
	}

	r.logger.Info("created shopping cart", zap.String("cartID", cartID.ID.String()))

	price, err := strconv.ParseFloat(cartID.TotalPrice, 64)
	if err != nil {
		return model.ShoppingCart{}, fmt.Errorf("parse price: %w", err)
	}

	return model.ShoppingCart{
		ID:         cartID.ID,
		UserID:     cartID.UserID.UUID,
		SessionID:  cartID.SessionID.UUID,
		TotalItems: cartID.TotalItems,
		TotalPrice: price,
	}, nil

}

// GetShoppingCartByUserID retrieves the shopping cart for a user
func (r *CartRepository) GetShoppingCartByUserID(ctx context.Context, userID uuid.NullUUID) (model.ShoppingCart, error) {
	cart, err := r.Queries.GetShoppingCartByUserID(ctx, userID)
	if err != nil {
		return model.ShoppingCart{}, fmt.Errorf("get shopping cart by user ID: %w", err)
	}

	price, err := strconv.ParseFloat(cart.TotalPrice, 64)
	if err != nil {
		return model.ShoppingCart{}, fmt.Errorf("parse price: %w", err)
	}

	return model.ShoppingCart{
		ID:         cart.ID,
		UserID:     cart.UserID.UUID,
		SessionID:  cart.SessionID.UUID,
		TotalItems: cart.TotalItems,
		TotalPrice: price,
	}, nil
}

// GetShoppingCartBySessionID retrieves the shopping cart for a session
func (r *CartRepository) GetShoppingCartBySessionID(ctx context.Context, sessionID uuid.NullUUID) (model.ShoppingCart, error) {
	cart, err := r.Queries.GetShoppingCartBySessionID(ctx, sessionID)
	if err != nil {
		return model.ShoppingCart{}, fmt.Errorf("get shopping cart by session ID: %w", err)
	}

	price, err := strconv.ParseFloat(cart.TotalPrice, 64)
	if err != nil {
		return model.ShoppingCart{}, fmt.Errorf("parse price: %w", err)
	}

	return model.ShoppingCart{
		ID:         cart.ID,
		UserID:     cart.UserID.UUID,
		SessionID:  cart.SessionID.UUID,
		TotalItems: cart.TotalItems,
		TotalPrice: price,
	}, nil
}

// DeleteShoppingCart deletes a shopping cart
func (r *CartRepository) DeleteShoppingCart(ctx context.Context, cartID uuid.UUID) error {
	return r.execTx(ctx, func(q *database.Queries) error {
		err := q.DeleteShoppingCart(ctx, cartID)
		if err != nil {
			return fmt.Errorf("delete shopping cart: %w", err)
		}

		r.logger.Info("deleted shopping cart", zap.String("cartID", cartID.String()))
		return nil
	})
}

// UpdateCartTotals updates the total price and number of items in the cart
func (r *CartRepository) UpdateCartTotals(ctx context.Context, cartID uuid.UUID) error {
	return r.execTx(ctx, func(q *database.Queries) error {
		err := q.UpdateCartTotals(ctx, cartID)
		if err != nil {
			return fmt.Errorf("update cart totals: %w", err)
		}

		r.logger.Info("updated cart totals", zap.String("cartID", cartID.String()))
		return nil
	})
}
