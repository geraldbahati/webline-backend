package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/middleware"
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

type contextKey string

const txKey contextKey = "tx"

// NewCartService creates a new CartService.
func NewCartService(
	logger *zap.Logger,
	config *appconfig.Config,
	cartRepository repository.CartRepository,
	productRepository *repository.ProductRepository,
	productImageRepository *repository.ProductImageRepository,
	cacheService CacheService,
) *CartService {
	return &CartService{
		logger:                 logger,
		config:                 config,
		cartRepository:         cartRepository,
		productRepository:      productRepository,
		productImageRepository: productImageRepository,
		cacheService:           cacheService,
	}
}

// constructS3URL constructs the S3 URL for a given file path
func (s *CartService) constructS3URL(filePath string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.AWSBucketName, s.config.AWSRegion, filePath)
}

// getOrCreateCart gets an existing cart or creates a new one based on session/user context
func (s *CartService) getOrCreateCart(ctx context.Context, session model.Session, userType string) (*model.ShoppingCart, error) {
	var cart *model.ShoppingCart
	var err error

	switch userType {
	case middleware.UserTypeAuthenticated:
		if session.UserID == nil {
			return nil, fmt.Errorf("authenticated session without user ID")
		}
		// Try to get cart by user ID
		cart, err = s.cartRepository.GetShoppingCartByUserID(ctx, *session.UserID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get user cart: %w", err)
		}

		if errors.Is(err, sql.ErrNoRows) {
			// Create new cart for authenticated user
			cart, err = s.cartRepository.CreateShoppingCart(ctx, repository.CreateShoppingCartParams{
				UserID: session.UserID,
			})
			if err != nil {
				return nil, fmt.Errorf("create user cart: %w", err)
			}
		}

	case middleware.UserTypeGuest:
		// Try to get cart by session ID
		cart, err = s.cartRepository.GetCartByGuestID(ctx, session.SessionID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get guest cart: %w", err)
		}

		if errors.Is(err, sql.ErrNoRows) {
			// Create new cart for guest
			cart, err = s.cartRepository.CreateShoppingCart(ctx, repository.CreateShoppingCartParams{
				GuestID: &session.SessionID,
			})
			if err != nil {
				return nil, fmt.Errorf("create guest cart: %w", err)
			}
		}

	default:
		return nil, fmt.Errorf("invalid user type: %s", userType)
	}

	return cart, nil
}

// AddToCart adds an item to the cart
func (s *CartService) AddToCart(ctx context.Context, session model.Session, userType string, productID string, quantity int32) error {
	cart, err := s.getOrCreateCart(ctx, session, userType)
	if err != nil {
		return fmt.Errorf("get or create cart: %w", err)
	}

	// Parse product ID
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return fmt.Errorf("invalid product ID: %w", err)
	}

	// Get product details
	product, err := s.productRepository.GetProductByID(ctx, productUUID)
	if err != nil {
		return fmt.Errorf("get product: %w", err)
	}

	// Add item to cart
	cartItem, err := s.cartRepository.UpsertCartItem(ctx, cart.ID, productUUID, quantity, product.USD)
	if err != nil {
		return fmt.Errorf("upsert cart item: %w", err)
	}

	// Update cache
	if err := s.updateCartCache(ctx, cart.ID, cartItem); err != nil {
		s.logger.Warn("Failed to update cart cache", zap.Error(err))
	}

	return nil
}

// GetCartItems returns all items in the cart
func (s *CartService) GetCartItems(ctx context.Context, session model.Session, userType string) ([]*model.CartItem, error) {
	var ownerUUID uuid.UUID
	var err error

	switch userType {
	case middleware.UserTypeAuthenticated:
		ownerUUID, err = uuid.Parse(session.UserID.String())
		if err != nil {
			s.logger.Error("Invalid owner ID format", zap.Error(err))
			return nil, fmt.Errorf("invalid owner ID format: %w", err)
		}
	case middleware.UserTypeGuest:
		ownerUUID = session.SessionID
	default:
		return nil, fmt.Errorf("invalid user type: %s", userType)
	}

	// Try to get cart
	cart, err := s.cartRepository.GetCartByOwnerID(ctx, ownerUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return empty list for new cart
			return []*model.CartItem{}, nil
		}
		s.logger.Error("Failed to get cart", zap.Error(err))
		return nil, fmt.Errorf("get cart: %w", err)
	}

	// Try cache first
	var items []*model.CartItem
	cacheKey := CartItemsKey(cart.ID.String())
	err = s.cacheService.Get(ctx, cacheKey, &items)
	if err == nil {
		return items, nil
	}

	// Get from database
	items, err = s.cartRepository.GetAllCartItems(ctx, cart.ID)
	if err != nil {
		s.logger.Error("Failed to get cart items", zap.Error(err))
		return nil, fmt.Errorf("get cart items: %w", err)
	}

	// Update cache
	if err := s.cacheService.SetWithTTL(ctx, cacheKey, items, 24*time.Hour); err != nil {
		s.logger.Warn("Failed to cache cart items", zap.Error(err))
	}

	return items, nil
}

// updateCartCache helper function to manage cart caching
func (s *CartService) updateCartCache(ctx context.Context, cartID uuid.UUID, item *model.CartItem) error {
	// Update item cache
	itemKey := CartItemKey(cartID.String(), item.ProductID.String())
	if err := s.cacheService.SetWithTTL(ctx, itemKey, item, 24*time.Hour); err != nil {
		return err
	}

	// Invalidate items list cache
	itemsKey := CartItemsKey(cartID.String())
	if err := s.cacheService.Delete(ctx, itemsKey); err != nil {
		return err
	}

	// Invalidate total cache
	totalKey := CartTotalKey(cartID.String())
	if err := s.cacheService.Delete(ctx, totalKey); err != nil {
		return err
	}

	return nil
}

// MigrateGuestCart migrates a guest cart to a user cart after login
func (s *CartService) MigrateGuestCart(ctx context.Context, guestSession model.Session, userID uuid.UUID) error {
	// Get guest cart
	guestCart, err := s.cartRepository.GetCartByGuestID(ctx, guestSession.SessionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("get guest cart: %w", err)
	}

	if guestCart == nil {
		// No guest cart to migrate
		return nil
	}

	// Get or create user cart
	userCart, err := s.cartRepository.GetShoppingCartByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			userCart, err = s.cartRepository.CreateShoppingCart(ctx, repository.CreateShoppingCartParams{
				UserID: &userID,
			})
			if err != nil {
				return fmt.Errorf("create user cart: %w", err)
			}
		} else {
			return fmt.Errorf("get user cart: %w", err)
		}
	}

	// Begin transaction for migration
	tx, err := s.cartRepository.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get all items from guest cart
	items, err := s.cartRepository.GetAllCartItems(ctx, guestCart.ID)
	if err != nil {
		return fmt.Errorf("get guest cart items: %w", err)
	}

	// Move items to user cart
	for _, item := range items {
		_, err = s.cartRepository.UpsertCartItem(ctx, userCart.ID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			return fmt.Errorf("migrate cart item: %w", err)
		}
	}

	// Delete guest cart
	if err := s.cartRepository.DeleteShoppingCart(ctx, guestCart.ID); err != nil {
		return fmt.Errorf("delete guest cart: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// Clear cache for both carts
	s.clearCartCache(ctx, guestCart.ID)
	s.clearCartCache(ctx, userCart.ID)

	return nil
}

// Helper function to clear cart cache
func (s *CartService) clearCartCache(ctx context.Context, cartID uuid.UUID) {
	cacheKeys := []string{
		CartItemsKey(cartID.String()),
		CartTotalKey(cartID.String()),
	}
	for _, key := range cacheKeys {
		if err := s.cacheService.Delete(ctx, key); err != nil {
			s.logger.Warn("Failed to delete cart cache",
				zap.Error(err),
				zap.String("key", key))
		}
	}
}

// RemoveFromCart removes an item from the cart.
func (s *CartService) RemoveFromCart(ctx context.Context, session model.Session, userType string, productID string) error {
	cart, err := s.getOrCreateCart(ctx, session, userType)
	if err != nil {
		return err
	}

	cartUUID, err := uuid.Parse(cart.ID.String())
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

	// Remove from cache for the specific cart item
	cacheKey := CartItemKey(cart.ID.String(), productID)
	err = s.cacheService.Delete(ctx, cacheKey)
	if err != nil {
		s.logger.Warn("Failed to delete cart item from cache", zap.Error(err))
	}

	// Invalidate the cached list of cart items
	cartItemsCacheKey := CartItemsKey(cart.ID.String())
	if delErr := s.cacheService.Delete(ctx, cartItemsCacheKey); delErr != nil {
		s.logger.Warn("Failed to delete cart items cache", zap.String("key", cartItemsCacheKey), zap.Error(delErr))
	}

	s.logger.Info("Removed product from cart", zap.String("productID", productID), zap.String("cartID", cart.ID.String()))
	return nil
}

// ClearCart removes all items from the cart.
func (s *CartService) ClearCart(ctx context.Context, session model.Session, userType string) error {
	cart, err := s.getOrCreateCart(ctx, session, userType)
	if err != nil {
		return err
	}

	cartUUID, err := uuid.Parse(cart.ID.String())
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
		CartItemsKey(cart.ID.String()),
		CartTotalKey(cart.ID.String()),
	}
	for _, key := range cacheKeys {
		if delErr := s.cacheService.Delete(ctx, key); delErr != nil {
			s.logger.Warn("Failed to delete cache key", zap.String("key", key), zap.Error(delErr))
		}
	}

	s.logger.Info("Cleared cart", zap.String("cartID", cart.ID.String()))
	return nil
}

// CalculateCartTotal calculates the total price of all items in the cart.
func (s *CartService) CalculateCartTotal(ctx context.Context, session model.Session, userType string) (float64, error) {
	cart, err := s.getOrCreateCart(ctx, session, userType)
	if err != nil {
		return 0, err
	}

	cartUUID, err := uuid.Parse(cart.ID.String())
	if err != nil {
		s.logger.Error("Failed to parse cart ID", zap.Error(err))
		return 0, fmt.Errorf("parse cart ID: %w", err)
	}

	// Attempt to retrieve total from cache.
	cacheKey := CartTotalKey(cart.ID.String())
	var total float64
	err = s.cacheService.GetOrSet(ctx, cacheKey, &total, func() error {
		// Calculate total using repository.
		t, err := s.cartRepository.CalculateCartTotal(ctx, cartUUID)
		if err != nil {
			return err
		}
		total = t
		return nil
	})

	if err != nil {
		s.logger.Error("Failed to calculate cart total", zap.Error(err))
		return 0, fmt.Errorf("calculate cart total: %w", err)
	}

	return total, nil
}

// UpdateCartItemQuantity updates the quantity of an item in the cart.
func (s *CartService) UpdateCartItemQuantity(ctx context.Context, session model.Session, userType string, productID string, quantity int32) error {
	cart, err := s.getOrCreateCart(ctx, session, userType)
	if err != nil {
		return err
	}

	cartUUID, err := uuid.Parse(cart.ID.String())
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

	// Update the cart item's quantity in the repository.
	err = s.cartRepository.UpdateCartItemQuantity(ctx, cartUUID, productUUID, quantity)
	if err != nil {
		s.logger.Error("Failed to update cart item quantity", zap.Error(err))
		return fmt.Errorf("update cart item quantity: %w", err)
	}

	// Fetch the updated cart item from the repository.
	updatedItem, err := s.cartRepository.GetCartItem(ctx, cartUUID, productUUID)
	if err != nil {
		s.logger.Warn("Failed to fetch updated cart item", zap.Error(err))
	} else {
		// Construct the S3 URL for the image
		if updatedItem.ImageURL != "" {
			imagePath := updatedItem.ImageURL
			s3URL := s.constructS3URL(imagePath)
			updatedItem.ImageURL = s3URL
		}

		// Update the cache with the new cart item.
		cacheKey := CartItemKey(cart.ID.String(), productID)
		err = s.cacheService.Set(ctx, cacheKey, *updatedItem)
		if err != nil {
			s.logger.Warn("Failed to update cart item cache", zap.Error(err))
		}
	}

	// Invalidate the cached list of cart items
	cartItemsCacheKey := CartItemsKey(cart.ID.String())
	if delErr := s.cacheService.Delete(ctx, cartItemsCacheKey); delErr != nil {
		s.logger.Warn("Failed to delete cart items cache", zap.String("key", cartItemsCacheKey), zap.Error(delErr))
	}

	// Invalidate the cart total cache as it has changed.
	totalCacheKey := CartTotalKey(cart.ID.String())
	if delErr := s.cacheService.Delete(ctx, totalCacheKey); delErr != nil {
		s.logger.Warn("Failed to delete cart total cache", zap.String("key", totalCacheKey), zap.Error(delErr))
	}

	s.logger.Info("Updated cart item quantity", zap.String("productID", productID), zap.String("cartID", cart.ID.String()), zap.Int32("quantity", quantity))
	return nil

}

// CreateShoppingCart creates a new shopping cart associated with a user or session.
func (s *CartService) CreateShoppingCart(ctx context.Context, session model.Session, userType string) (model.ShoppingCart, error) {
	// Create the shopping cart in the repository.
	cart, err := s.getOrCreateCart(ctx, session, userType)
	if err != nil {
		return model.ShoppingCart{}, err
	}

	return *cart, nil
}

// ReplaceCartItems replaces all items in the cart with the provided items.
func (s *CartService) ReplaceCartItems(ctx context.Context, session model.Session, userType string, items []CartItemInput) error {
	cart, err := s.getOrCreateCart(ctx, session, userType)
	if err != nil {
		return err
	}

	cartUUID, err := uuid.Parse(cart.ID.String())
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
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.logger.Error("Failed to rollback transaction", zap.Error(rollbackErr))
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				s.logger.Error("Failed to commit transaction", zap.Error(commitErr))
			}
		}
	}()

	txCtx := context.WithValue(ctx, txKey, tx)

	// Clear existing cart items.
	if err := s.cartRepository.ClearCart(txCtx, cartUUID); err != nil {
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

		product, err := s.productRepository.GetProductByID(txCtx, productUUID)
		if err != nil {
			s.logger.Error("Failed to get product by ID", zap.Error(err))
			return fmt.Errorf("get product by ID: %w", err)
		}

		cartItem, err := s.cartRepository.UpsertCartItem(txCtx, cartUUID, product.ID, item.Quantity, fmt.Sprintf("%.2f", item.Price))
		if err != nil {
			s.logger.Error("Failed to upsert cart item", zap.Error(err))
			return fmt.Errorf("upsert cart item: %w", err)
		}

		// Construct the S3 URL for the image
		if cartItem.ImageURL != "" {
			imagePath := cartItem.ImageURL
			s3URL := s.constructS3URL(imagePath)
			cartItem.ImageURL = s3URL
		}

		// Update cache for each item.
		cacheKey := CartItemKey(cart.ID.String(), item.ProductID)
		err = s.cacheService.Set(ctx, cacheKey, *cartItem)
		if err != nil {
			s.logger.Warn("Failed to cache cart item", zap.Error(err))
		}
	}

	// Invalidate the cached list of cart items
	cartItemsCacheKey := CartItemsKey(cart.ID.String())
	if err := s.cacheService.Delete(ctx, cartItemsCacheKey); err != nil {
		s.logger.Warn("Failed to delete cart items cache", zap.String("key", cartItemsCacheKey), zap.Error(err))
	}

	// Invalidate the cart total cache.
	totalCacheKey := CartTotalKey(cart.ID.String())
	if err := s.cacheService.Delete(ctx, totalCacheKey); err != nil {
		s.logger.Warn("Failed to delete cart total cache", zap.String("key", totalCacheKey), zap.Error(err))
	}

	s.logger.Info("Replaced cart items", zap.String("cartID", cart.ID.String()))
	return nil
}

// GetShoppingCart retrieves the shopping cart by its ID.
func (s *CartService) GetShoppingCart(ctx context.Context, session model.Session, userType string) (model.ShoppingCart, error) {
	cart, err := s.getOrCreateCart(ctx, session, userType)
	if err != nil {
		return model.ShoppingCart{}, err
	}

	return *cart, nil
}
