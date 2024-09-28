// services/cart_service.go
package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CartService provides methods to manage shopping carts.
type CartService struct {
	logger                 *zap.Logger
	config                 *appconfig.Config
	cartRepository         repository.CartRepository
	productRepository      *repository.ProductRepository
	productImageRepository *repository.ProductImageRepository
	cacheService           CacheService
}

// CartItemInput represents the input for adding/updating a cart item.
type CartItemInput struct {
	ProductID string
	Quantity  int32
	Price     float64
}

// NewCartService creates a new CartService.
func NewCartService(logger *zap.Logger, config *appconfig.Config, cartRepository repository.CartRepository, productRepository *repository.ProductRepository, productImageRepository *repository.ProductImageRepository, cacheService CacheService) *CartService {
	return &CartService{
		logger:                 logger,
		config:                 config,
		cartRepository:         cartRepository,
		productRepository:      productRepository,
		productImageRepository: productImageRepository,
		cacheService:           cacheService,
	}
}

// AddToCart adds an item to the cart or updates the quantity if it already exists.
func (s *CartService) AddToCart(ctx context.Context, cartID, productID string, quantity int32, price float64) error {
	cartUUID, err := uuid.Parse(cartID)
	if err != nil {
		s.logger.Error("Failed to parse cart ID", zap.Error(err))
		return fmt.Errorf("parse cart ID: %w", err)
	}

	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("Failed to parse product ID", zap.Error(err))
		return fmt.Errorf("parse product ID: %w", err)
	}

	// Retrieve product by ID
	product, err := s.productRepository.GetProductByID(ctx, productUUID)
	if err != nil {
		s.logger.Error("Failed to get product by ID", zap.Error(err))
		return fmt.Errorf("get product by ID: %w", err)
	}

	// Create or update the cart item
	cartItem, err := s.cartRepository.UpsertCartItem(ctx, cartUUID, product.ID, quantity, fmt.Sprintf("%.2f", price))
	if err != nil {
		s.logger.Error("Failed to upsert cart item", zap.Error(err))
		return fmt.Errorf("upsert cart item: %w", err)
	}

	// Update cache
	cacheKey := fmt.Sprintf("cart:%s:item:%s", cartID, productID)
	err = s.cacheService.Set(ctx, cacheKey, *cartItem)
	if err != nil {
		s.logger.Warn("Failed to cache cart item", zap.Error(err))
	}

	s.logger.Info("Added/Updated product in cart", zap.String("productID", product.ID.String()), zap.String("cartID", cartID))
	return nil
}

// RemoveFromCart removes an item from the cart.
func (s *CartService) RemoveFromCart(ctx context.Context, cartID, productID string) error {
	cartUUID, err := uuid.Parse(cartID)
	if err != nil {
		s.logger.Error("Failed to parse cart ID", zap.Error(err))
		return fmt.Errorf("parse cart ID: %w", err)
	}

	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("Failed to parse product ID", zap.Error(err))
		return fmt.Errorf("parse product ID: %w", err)
	}

	// Remove the cart item
	err = s.cartRepository.RemoveCartItem(ctx, cartUUID, productUUID)
	if err != nil {
		s.logger.Error("Failed to remove cart item", zap.Error(err))
		return fmt.Errorf("remove cart item: %w", err)
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("cart:%s:item:%s", cartID, productID)
	err = s.cacheService.Delete(ctx, cacheKey)
	if err != nil {
		s.logger.Warn("Failed to delete cart item from cache", zap.Error(err))
	}

	s.logger.Info("Removed product from cart", zap.String("productID", productID), zap.String("cartID", cartID))
	return nil
}

// GetCartItems returns all items in the cart.
func (s *CartService) GetCartItems(ctx context.Context, cartID string) ([]model.CartItem, error) {
	cartUUID, err := uuid.Parse(cartID)
	if err != nil {
		s.logger.Error("Failed to parse cart ID", zap.Error(err))
		return nil, fmt.Errorf("parse cart ID: %w", err)
	}

	// Attempt to retrieve cart items from cache.
	cacheKey := fmt.Sprintf("cart:%s:items", cartID)
	var cartItems []model.CartItem
	err = s.cacheService.GetOrSet(ctx, cacheKey, &cartItems, func() error {
		items, err := s.cartRepository.GetAllCartItems(ctx, cartUUID)
		if err != nil {
			return err
		}
		cartItems = items
		return nil
	})

	if err != nil {
		s.logger.Error("Failed to get cart items", zap.Error(err))
		return nil, fmt.Errorf("get cart items: %w", err)
	}

	return cartItems, nil
}

// ClearCart removes all items from the cart.
func (s *CartService) ClearCart(ctx context.Context, cartID string) error {
	cartUUID, err := uuid.Parse(cartID)
	if err != nil {
		s.logger.Error("Failed to parse cart ID", zap.Error(err))
		return fmt.Errorf("parse cart ID: %w", err)
	}

	// Clear the cart in the repository.
	err = s.cartRepository.ClearCart(ctx, cartUUID)
	if err != nil {
		s.logger.Error("Failed to clear cart", zap.Error(err))
		return fmt.Errorf("clear cart: %w", err)
	}

	// Invalidate cache related to the cart.
	cacheKeys := []string{
		fmt.Sprintf("cart:%s:items", cartID),
		fmt.Sprintf("cart:%s:total", cartID),
	}
	for _, key := range cacheKeys {
		if delErr := s.cacheService.Delete(ctx, key); delErr != nil {
			s.logger.Warn("Failed to delete cache key", zap.String("key", key), zap.Error(delErr))
		}
	}

	s.logger.Info("Cleared cart", zap.String("cartID", cartID))
	return nil
}

// CalculateCartTotal calculates the total price of all items in the cart.
func (s *CartService) CalculateCartTotal(ctx context.Context, cartID string) (float64, error) {
	cartUUID, err := uuid.Parse(cartID)
	if err != nil {
		s.logger.Error("Failed to parse cart ID", zap.Error(err))
		return 0, fmt.Errorf("parse cart ID: %w", err)
	}

	// Attempt to retrieve total from cache.
	cacheKey := fmt.Sprintf("cart:%s:total", cartID)
	var total float64
	err = s.cacheService.GetOrSet(ctx, cacheKey, &total, func() error {
		// Calculate total using repository.
		t, err := s.cartRepository.CalculateCartTotal(ctx, cartUUID)
		if err != nil {
			return err
		}
		total = float64(t)
		return nil
	})

	if err != nil {
		s.logger.Error("Failed to calculate cart total", zap.Error(err))
		return 0, fmt.Errorf("calculate cart total: %w", err)
	}

	return total, nil
}

// UpdateCartItemQuantity updates the quantity of an item in the cart.
func (s *CartService) UpdateCartItemQuantity(ctx context.Context, cartID, productID string, quantity int32) error {
	cartUUID, err := uuid.Parse(cartID)
	if err != nil {
		s.logger.Error("Failed to parse cart ID", zap.Error(err))
		return fmt.Errorf("parse cart ID: %w", err)
	}

	if quantity < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}

	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("Failed to parse product ID", zap.Error(err))
		return fmt.Errorf("parse product ID: %w", err)
	}

	// Update the cart item's quantity.
	err = s.cartRepository.UpdateCartItemQuantity(ctx, cartUUID, productUUID, quantity)
	if err != nil {
		s.logger.Error("Failed to update cart item quantity", zap.Error(err))
		return fmt.Errorf("update cart item quantity: %w", err)
	}

	// Update the cache for the specific cart item.
	cacheKey := fmt.Sprintf("cart:%s:item:%s", cartID, productID)
	var updatedItem model.CartItem
	err = s.cacheService.GetOrSet(ctx, cacheKey, &updatedItem, func() error {
		item, err := s.cartRepository.GetCartItem(ctx, cartUUID, productUUID)
		if err != nil {
			return err
		}
		updatedItem = *item
		return nil
	})

	if err != nil {
		s.logger.Warn("Failed to update cart item cache", zap.Error(err))
	}

	// Invalidate the cart total cache as it has changed.
	totalCacheKey := fmt.Sprintf("cart:%s:total", cartID)
	if delErr := s.cacheService.Delete(ctx, totalCacheKey); delErr != nil {
		s.logger.Warn("Failed to delete cart total cache", zap.String("key", totalCacheKey), zap.Error(delErr))
	}

	s.logger.Info("Updated cart item quantity", zap.String("productID", productID), zap.String("cartID", cartID), zap.Int32("quantity", quantity))
	return nil
}

// CreateShoppingCart creates a new shopping cart associated with a session or user.
func (s *CartService) CreateShoppingCart(ctx context.Context, sessionID *uuid.UUID, userID *uuid.UUID) (model.ShoppingCart, error) {
	// Create the shopping cart in the repository.
	cart, err := s.cartRepository.CreateShoppingCart(ctx, userID, sessionID)
	if err != nil {
		s.logger.Error("Failed to create shopping cart", zap.Error(err))
		return model.ShoppingCart{}, fmt.Errorf("create shopping cart: %w", err)
	}

	// Optionally, you can set initial totals or perform other initializations here.

	s.logger.Info("Created shopping cart", zap.String("cartID", cart.ID.String()))
	return *cart, nil
}

// GetShoppingCartByUserID retrieves the shopping cart associated with a user.
func (s *CartService) GetShoppingCartByUserID(ctx context.Context, userID string) (string, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Error("Failed to parse user ID", zap.Error(err))
		return "", fmt.Errorf("parse user ID: %w", err)
	}

	cart, err := s.cartRepository.GetShoppingCartByUserID(ctx, userUUID)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err == sql.ErrNoRows)
		s.logger.Error("Failed to get shopping cart by user ID", zap.Error(err))
		return "", fmt.Errorf("get shopping cart by user ID: %w", err)
	}

	if err == sql.ErrNoRows {
		// Create a new shopping cart
		cart, err := s.CreateShoppingCart(ctx, nil, &userUUID)
		if err != nil {
			s.logger.Error("Failed to create shopping cart", zap.Error(err))
			return "", fmt.Errorf("create shopping cart: %w", err)
		}
		return cart.ID.String(), nil
	}

	return cart.ID.String(), nil
}

// GetShoppingCartBySessionID retrieves the shopping cart associated with a session.
func (s *CartService) GetShoppingCartBySessionID(ctx context.Context, sessionID string) (string, error) {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		s.logger.Error("Failed to parse session ID", zap.Error(err))
		return "", fmt.Errorf("parse session ID: %w", err)
	}

	cart, err := s.cartRepository.GetCartBySessionReference(ctx, sessionUUID)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Error("Failed to get shopping cart by session ID", zap.Error(err))
		return "", fmt.Errorf("get shopping cart by session ID: %w", err)
	}

	if err == sql.ErrNoRows {
		// Create a new shopping cart
		cart, err := s.CreateShoppingCart(ctx, &sessionUUID, nil)
		if err != nil {
			s.logger.Error("Failed to create shopping cart", zap.Error(err))
			return "", fmt.Errorf("create shopping cart: %w", err)
		}
		return cart.ID.String(), nil
	}

	return cart.ID.String(), nil
}

// ReplaceCartItems replaces all items in the cart with the provided items.
func (s *CartService) ReplaceCartItems(ctx context.Context, cartID string, items []CartItemInput) error {
	cartUUID, err := uuid.Parse(cartID)
	if err != nil {
		s.logger.Error("Failed to parse cart ID", zap.Error(err))
		return fmt.Errorf("parse cart ID: %w", err)
	}

	// Begin a transaction to ensure atomicity.
	tx, err := s.cartRepository.BeginTx(ctx)
	if err != nil {
		s.logger.Error("Failed to begin transaction", zap.Error(err))
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Clear existing cart items.
	if err := s.cartRepository.ClearCart(ctx, cartUUID); err != nil {
		s.logger.Error("Failed to clear cart items", zap.Error(err))
		return fmt.Errorf("clear cart items: %w", err)
	}

	// Add new cart items.
	for _, item := range items {
		productUUID, err := uuid.Parse(item.ProductID)
		if err != nil {
			s.logger.Error("Failed to parse product ID", zap.Error(err))
			return fmt.Errorf("parse product ID: %w", err)
		}

		product, err := s.productRepository.GetProductByID(ctx, productUUID)
		if err != nil {
			s.logger.Error("Failed to get product by ID", zap.Error(err))
			return fmt.Errorf("get product by ID: %w", err)
		}

		_, err = s.cartRepository.UpsertCartItem(ctx, cartUUID, product.ID, item.Quantity, fmt.Sprintf("%.2f", item.Price))
		if err != nil {
			s.logger.Error("Failed to upsert cart item", zap.Error(err))
			return fmt.Errorf("upsert cart item: %w", err)
		}

		// Update cache for each item.
		cacheKey := fmt.Sprintf("cart:%s:item:%s", cartID, item.ProductID)
		err = s.cacheService.Set(ctx, cacheKey, model.CartItem{
			ProductID: product.ID,
			Quantity:  item.Quantity,
			Price:     fmt.Sprintf("%.2f", item.Price),
		})
		if err != nil {
			s.logger.Warn("Failed to cache cart item", zap.Error(err))
		}
	}

	// Invalidate the cart total cache.
	totalCacheKey := fmt.Sprintf("cart:%s:total", cartID)
	if err := s.cacheService.Delete(ctx, totalCacheKey); err != nil {
		s.logger.Warn("Failed to delete cart total cache", zap.String("key", totalCacheKey), zap.Error(err))
	}

	s.logger.Info("Replaced cart items", zap.String("cartID", cartID))
	return nil
}
