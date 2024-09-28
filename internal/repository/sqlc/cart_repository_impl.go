package sqlc

import (
	"context"
	"database/sql"
	"fmt"
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

func (r *cartRepositoryImpl) CreateShoppingCart(ctx context.Context, userID, sessionReference *uuid.UUID) (*model.ShoppingCart, error) {
	var userNull, sessionNull uuid.NullUUID
	if userID != nil {
		userNull = uuid.NullUUID{UUID: *userID, Valid: true}
	}
	if sessionReference != nil {
		sessionNull = uuid.NullUUID{UUID: *sessionReference, Valid: true}
	}

	params := database.CreateShoppingCartParams{
		UserID:           userNull,
		SessionReference: sessionNull,
	}
	cart, err := r.Queries.CreateShoppingCart(ctx, params)
	if err != nil {
		r.logger.Error("failed to create shopping cart", zap.Error(err))
		return nil, fmt.Errorf("create shopping cart: %w", err)
	}
	return convertDBShoppingCartToModel(cart), nil
}

func (r *cartRepositoryImpl) GetShoppingCartByUserID(ctx context.Context, userID uuid.UUID) (*model.ShoppingCart, error) {
	cart, err := r.Queries.GetShoppingCartByUserID(ctx, uuid.NullUUID{UUID: userID, Valid: true})
	if err != nil {
		r.logger.Error("failed to get shopping cart by user ID", zap.Error(err))
		return nil, err
	}
	return convertDBShoppingCartToModel(cart), nil
}

func (r *cartRepositoryImpl) GetCartBySessionReference(ctx context.Context, sessionReference uuid.UUID) (*model.ShoppingCart, error) {
	cart, err := r.Queries.GetCartBySessionReference(ctx, uuid.NullUUID{UUID: sessionReference, Valid: true})
	if err != nil {
		r.logger.Error("failed to get cart by session reference", zap.Error(err))
		return nil, err
	}
	return convertDBShoppingCartToModel(cart), nil
}

func (r *cartRepositoryImpl) DeleteShoppingCart(ctx context.Context, id uuid.UUID) error {
	err := r.Queries.DeleteShoppingCart(ctx, id)
	if err != nil {
		r.logger.Error("failed to delete shopping cart", zap.Error(err))
		return fmt.Errorf("delete shopping cart: %w", err)
	}
	return nil
}

func (r *cartRepositoryImpl) UpdateCartTotals(ctx context.Context, id uuid.UUID) error {
	err := r.Queries.UpdateCartTotals(ctx, id)
	if err != nil {
		r.logger.Error("failed to update cart totals", zap.Error(err))
		return fmt.Errorf("update cart totals: %w", err)
	}
	return nil
}

func (r *cartRepositoryImpl) CalculateCartTotal(ctx context.Context, shoppingCartID uuid.UUID) (int64, error) {
	total, err := r.Queries.CalculateCartTotal(ctx, shoppingCartID)
	if err != nil {
		r.logger.Error("failed to calculate cart total", zap.Error(err))
		return 0, fmt.Errorf("calculate cart total: %w", err)
	}
	return total, nil
}

func (r *cartRepositoryImpl) ClearCart(ctx context.Context, shoppingCartID uuid.UUID) error {
	err := r.Queries.ClearCart(ctx, shoppingCartID)
	if err != nil {
		r.logger.Error("failed to clear cart", zap.Error(err))
		return fmt.Errorf("clear cart: %w", err)
	}
	return nil
}

func (r *cartRepositoryImpl) GetAllCartItems(ctx context.Context, shoppingCartID uuid.UUID) ([]model.CartItem, error) {
	items, err := r.Queries.GetAllCartItems(ctx, shoppingCartID)
	if err != nil {
		r.logger.Error("failed to get all cart items", zap.Error(err))
		return nil, fmt.Errorf("get all cart items: %w", err)
	}
	return convertDBCartItemsToModel(items), nil
}

func (r *cartRepositoryImpl) GetCartItem(ctx context.Context, shoppingCartID, productID uuid.UUID) (*model.CartItem, error) {
	item, err := r.Queries.GetCartItem(ctx, database.GetCartItemParams{
		ShoppingCartID: shoppingCartID,
		ProductID:      productID,
	})
	if err != nil {
		r.logger.Error("failed to get cart item", zap.Error(err))
		return nil, fmt.Errorf("get cart item: %w", err)
	}
	return convertDBCartItemToModel(item), nil
}

func (r *cartRepositoryImpl) RemoveCartItem(ctx context.Context, shoppingCartID, productID uuid.UUID) error {
	err := r.Queries.RemoveCartItem(ctx, database.RemoveCartItemParams{
		ShoppingCartID: shoppingCartID,
		ProductID:      productID,
	})
	if err != nil {
		r.logger.Error("failed to remove cart item", zap.Error(err))
		return fmt.Errorf("remove cart item: %w", err)
	}
	return nil
}

func (r *cartRepositoryImpl) UpdateCartItemQuantity(ctx context.Context, shoppingCartID, productID uuid.UUID, quantity int32) error {
	err := r.Queries.UpdateCartItemQuantity(ctx, database.UpdateCartItemQuantityParams{
		ShoppingCartID: shoppingCartID,
		ProductID:      productID,
		Quantity:       quantity,
	})
	if err != nil {
		r.logger.Error("failed to update cart item quantity", zap.Error(err))
		return fmt.Errorf("update cart item quantity: %w", err)
	}
	return nil
}

func (r *cartRepositoryImpl) UpdateCartUserID(ctx context.Context, cartID, userID uuid.UUID) error {
	err := r.Queries.UpdateCartUserID(ctx, database.UpdateCartUserIDParams{
		ID:     cartID,
		UserID: uuid.NullUUID{UUID: userID, Valid: true},
	})
	if err != nil {
		r.logger.Error("failed to update cart user ID", zap.Error(err))
		return fmt.Errorf("update cart user ID: %w", err)
	}
	return nil
}

func (r *cartRepositoryImpl) UpsertCartItem(ctx context.Context, shoppingCartID, productID uuid.UUID, quantity int32, price string) (*model.CartItem, error) {
	item, err := r.Queries.UpsertCartItem(ctx, database.UpsertCartItemParams{
		ShoppingCartID: shoppingCartID,
		ProductID:      productID,
		Quantity:       quantity,
		Price:          price,
	})
	if err != nil {
		r.logger.Error("failed to upsert cart item", zap.Error(err))
		return nil, fmt.Errorf("upsert cart item: %w", err)
	}
	return convertDBUpsertCartItemToModel(item), nil
}

// Helper functions to convert database models to domain models
func convertDBShoppingCartToModel(dbCart interface{}) *model.ShoppingCart {
	var cart model.ShoppingCart
	switch v := dbCart.(type) {
	case database.CreateShoppingCartRow:
		totalPrice, err := strconv.ParseFloat(v.TotalPrice, 64)
		if err != nil {
			return nil
		}
		cart = model.ShoppingCart{
			ID:         v.ID,
			UserID:     v.UserID.UUID,
			SessionID:  v.SessionReference.UUID,
			TotalItems: v.TotalItems,
			TotalPrice: totalPrice,
		}
	case database.GetShoppingCartByUserIDRow:
		totalPrice, err := strconv.ParseFloat(v.TotalPrice, 64)
		if err != nil {
			return nil
		}
		cart = model.ShoppingCart{
			ID:         v.ID,
			UserID:     v.UserID.UUID,
			SessionID:  v.SessionReference.UUID,
			TotalItems: v.TotalItems,
			TotalPrice: totalPrice,
		}
	case database.ShoppingCart:
		totalPrice, err := strconv.ParseFloat(v.TotalPrice, 64)
		if err != nil {
			return nil
		}
		cart = model.ShoppingCart{
			ID:         v.ID,
			UserID:     v.UserID.UUID,
			SessionID:  v.SessionReference.UUID,
			TotalItems: v.TotalItems,
			TotalPrice: totalPrice,
		}
	}
	return &cart
}

func convertDBCartItemsToModel(dbItems []database.GetAllCartItemsRow) []model.CartItem {
	items := make([]model.CartItem, len(dbItems))
	for i, dbItem := range dbItems {
		items[i] = model.CartItem{
			ID:        dbItem.ID,
			ProductID: dbItem.ProductID,
			Quantity:  dbItem.Quantity,
			Price:     dbItem.Price,
		}
	}
	return items
}

func convertDBCartItemToModel(dbItem database.GetCartItemRow) *model.CartItem {
	return &model.CartItem{
		ID:       dbItem.ID,
		Quantity: dbItem.Quantity,
	}
}

func convertDBUpsertCartItemToModel(dbItem database.UpsertCartItemRow) *model.CartItem {
	return &model.CartItem{
		ID:        dbItem.ID,
		ProductID: dbItem.ProductID,
		Quantity:  dbItem.Quantity,
		Price:     dbItem.Price,
	}
}
