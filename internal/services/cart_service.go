package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

// AddToCart adds an item to the cart or updates the quantity if it already exists.
func (s *CartService) AddToCart(ctx context.Context, user middleware.User, productID string, quantity int32, price float64) error {
	var cart *model.ShoppingCart
	var err error

	// Retrieve the shopping cart based on provided user information
	if user.IsGuest {
		// Handle guest user
		guestUUID, err := uuid.Parse(user.GuestID)
		if err != nil {
			s.logger.Error("Invalid guest ID format", zap.Error(err))
			return fmt.Errorf("invalid guest ID format: %w", err)
		}

		cart, err = s.cartRepository.GetCartByGuestID(ctx, guestUUID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Cart does not exist; create a new one
				cart, err = s.cartRepository.CreateShoppingCart(ctx, nil, &guestUUID)
				if err != nil {
					s.logger.Error("Failed to create shopping cart for guest", zap.Error(err))
					return fmt.Errorf("create shopping cart for guest: %w", err)
				}
			} else {
				s.logger.Error("Failed to get cart by guest ID", zap.Error(err))
				return fmt.Errorf("get cart by guest ID: %w", err)
			}
		}
	} else {
		// Handle authenticated user
		cart, err = s.cartRepository.GetShoppingCartByUserID(ctx, user.UserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Cart does not exist; create a new one
				cart, err = s.cartRepository.CreateShoppingCart(ctx, &user.UserID, nil)
				if err != nil {
					s.logger.Error("Failed to create shopping cart for user", zap.Error(err))
					return fmt.Errorf("create shopping cart for user: %w", err)
				}
			} else {
				s.logger.Error("Failed to get cart by user ID", zap.Error(err))
				return fmt.Errorf("get cart by user ID: %w", err)
			}
		}
	}

	// Parse productID
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

	// Upsert the cart item
	cartItem, err := s.cartRepository.UpsertCartItem(ctx, cart.ID, product.ID, quantity, fmt.Sprintf("%.2f", price))
	if err != nil {
		s.logger.Error("Failed to upsert cart item", zap.Error(err))
		return fmt.Errorf("upsert cart item: %w", err)
	}

	// Construct the S3 URL for the image (assuming constructS3URL is defined)
	if cartItem.ImageURL != "" {
		imagePath := cartItem.ImageURL
		s3URL := s.constructS3URL(imagePath)
		cartItem.ImageURL = s3URL
	}

	// Update cache for the specific cart item
	cacheKey := CartItemKey(cart.ID.String(), productID)
	err = s.cacheService.Set(ctx, cacheKey, *cartItem)
	if err != nil {
		s.logger.Warn("Failed to cache cart item", zap.Error(err))
	}

	// Invalidate the cached list of cart items
	cartItemsCacheKey := CartItemsKey(cart.ID.String())
	if delErr := s.cacheService.Delete(ctx, cartItemsCacheKey); delErr != nil {
		s.logger.Warn("Failed to delete cart items cache", zap.String("key", cartItemsCacheKey), zap.Error(delErr))
	}

	s.logger.Info("Added/Updated product in cart", zap.String("productID", product.ID.String()), zap.String("cartID", cart.ID.String()))
	return nil
}

// RemoveFromCart removes an item from the cart.
func (s *CartService) RemoveFromCart(ctx context.Context, user middleware.User, productID string) error {
	cart, err := s.getOrCreateCart(ctx, user)
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

// GetCartItems returns all items in the cart.
func (s *CartService) GetCartItems(ctx context.Context, user middleware.User) ([]*model.CartItem, error) {

	// Retrieve or create cart
	cart, err := s.getOrCreateCart(ctx, user)
	if err != nil {
		// Error already logged and wrapped in getOrCreateCart
		return nil, err
	}

	// Attempt to retrieve cart items from cache
	cacheKey := CartItemsKey(cart.ID.String())
	var cartItems []*model.CartItem
	err = s.cacheService.GetOrSet(ctx, cacheKey, &cartItems, func() error {
		items, err := s.cartRepository.GetAllCartItems(ctx, cart.ID)
		if err != nil {
			return fmt.Errorf("failed to get all cart items: %w", err)
		}

		// Optimize image URL construction
		for _, item := range items {
			if item.ImageURL != "" {
				item.ImageURL = s.constructS3URL(item.ImageURL)
			}
		}

		cartItems = items
		return nil
	})
	if err != nil {
		s.logger.Error("Failed to get cart items", zap.Error(err))
		return nil, fmt.Errorf("get cart items: %w", err)
	}

	s.logger.Info("Retrieved cart items", zap.String("cartID", cart.ID.String()), zap.Int("count", len(cartItems)))
	return cartItems, nil
}

func (s *CartService) getOrCreateCart(ctx context.Context, user middleware.User) (*model.ShoppingCart, error) {
	if user.IsGuest {
		// Handle guest user
		guestUUID, err := uuid.Parse(user.GuestID)
		if err != nil {
			s.logger.Error("Invalid guest ID format", zap.Error(err))
			return nil, fmt.Errorf("invalid guest ID format: %w", err)
		}

		cart, err := s.cartRepository.GetCartByGuestID(ctx, guestUUID)
		if err == nil {
			return cart, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			s.logger.Error("Failed to get cart by guest ID", zap.Error(err))
			return nil, fmt.Errorf("get cart by guest ID: %w", err)
		}

		// Create a new cart for guest
		cart, err = s.cartRepository.CreateShoppingCart(ctx, nil, &guestUUID)
		if err != nil {
			s.logger.Error("Failed to create shopping cart for guest", zap.Error(err))
			return nil, fmt.Errorf("create shopping cart for guest: %w", err)
		}
		return cart, nil
	}

	// Handle authenticated user
	cart, err := s.cartRepository.GetShoppingCartByUserID(ctx, user.UserID)
	if err == nil {
		return cart, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		s.logger.Error("Failed to get cart by user ID", zap.Error(err))
		return nil, fmt.Errorf("get cart by user ID: %w", err)
	}

	// Create a new cart for authenticated user
	cart, err = s.cartRepository.CreateShoppingCart(ctx, &user.UserID, nil)
	if err != nil {
		s.logger.Error("Failed to create shopping cart for user", zap.Error(err))
		return nil, fmt.Errorf("create shopping cart for user: %w", err)
	}
	return cart, nil
}

// ClearCart removes all items from the cart.
func (s *CartService) ClearCart(ctx context.Context, user middleware.User) error {
	cart, err := s.getOrCreateCart(ctx, user)
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
func (s *CartService) CalculateCartTotal(ctx context.Context, user middleware.User) (float64, error) {
	cart, err := s.getOrCreateCart(ctx, user)
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
func (s *CartService) UpdateCartItemQuantity(ctx context.Context, user middleware.User, productID string, quantity int32) error {
	cart, err := s.getOrCreateCart(ctx, user)
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
func (s *CartService) CreateShoppingCart(ctx context.Context, userID *uuid.UUID, guestID *uuid.UUID) (model.ShoppingCart, error) {
	// Create the shopping cart in the repository.
	cart, err := s.cartRepository.CreateShoppingCart(ctx, userID, guestID)
	if err != nil {
		s.logger.Error("Failed to create shopping cart", zap.Error(err))
		return model.ShoppingCart{}, fmt.Errorf("create shopping cart: %w", err)
	}

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
		s.logger.Error("Failed to get shopping cart by user ID", zap.Error(err))
		return "", fmt.Errorf("get shopping cart by user ID: %w", err)
	}

	if err == sql.ErrNoRows {
		// Create a new shopping cart
		cart, err := s.CreateShoppingCart(ctx, &userUUID, nil)
		if err != nil {
			s.logger.Error("Failed to create shopping cart", zap.Error(err))
			return "", fmt.Errorf("create shopping cart: %w", err)
		}
		return cart.ID.String(), nil
	}

	return cart.ID.String(), nil
}

// ReplaceCartItems replaces all items in the cart with the provided items.
func (s *CartService) ReplaceCartItems(ctx context.Context, user middleware.User, items []CartItemInput) error {
	cart, err := s.getOrCreateCart(ctx, user)
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
func (s *CartService) GetShoppingCart(ctx context.Context, user middleware.User) (model.ShoppingCart, error) {
	cart, err := s.getOrCreateCart(ctx, user)
	if err != nil {
		return model.ShoppingCart{}, err
	}

	return *cart, nil
}
