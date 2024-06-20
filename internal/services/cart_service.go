package services

import (
	"context"
	"fmt"
	"log"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CartService struct {
	logger                 *zap.Logger
	config                 *appconfig.Config
	cartRepository         *repository.CartRepository
	productRepository      *repository.ProductRepository
	productImageRepository *repository.ProductImageRepository
}

func NewCartService(logger *zap.Logger, config *appconfig.Config, cartRepository *repository.CartRepository, productRepository *repository.ProductRepository, productImageRepository *repository.ProductImageRepository) *CartService {
	return &CartService{
		logger:                 logger,
		config:                 config,
		cartRepository:         cartRepository,
		productRepository:      productRepository,
		productImageRepository: productImageRepository,
	}
}

// AddToCart adds an item to the cart or updates the quantity if it already exists
func (s *CartService) AddToCart(ctx context.Context, cartID, productID string, quantity int32, price float64) error {
	// Convert the string IDs to UUID
	cartUUID, err := uuid.Parse(cartID)
	if err != nil {
		s.logger.Error("failed to parse cart ID", zap.Error(err))
		return fmt.Errorf("parse cart ID: %w", err)
	}

	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("failed to parse product ID", zap.Error(err))
		return fmt.Errorf("parse product ID: %w", err)
	}

	cartNullUUID := uuid.NullUUID{UUID: cartUUID, Valid: true}
	productNullUUID := uuid.NullUUID{UUID: productUUID, Valid: true}

	log.Printf("Adding product with ID: %s to cart with ID: %s", productNullUUID.UUID.String(), cartNullUUID.UUID.String())

	return s.cartRepository.AddToCart(ctx, cartNullUUID, productNullUUID, quantity, price)
}

// RemoveFromCart removes an item from the cart
func (s *CartService) RemoveFromCart(ctx context.Context, cartID, productID string) error {
	// Convert the string IDs to UUID
	cartUUID, err := uuid.Parse(cartID)
	if err != nil {
		s.logger.Error("failed to parse cart ID", zap.Error(err))
		return fmt.Errorf("parse cart ID: %w", err)
	}

	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("failed to parse product ID", zap.Error(err))
		return fmt.Errorf("parse product ID: %w", err)
	}

	cartNullUUID := uuid.NullUUID{UUID: cartUUID, Valid: true}
	productNullUUID := uuid.NullUUID{UUID: productUUID, Valid: true}

	log.Printf("Removing product with ID: %s from cart with ID: %s", productNullUUID.UUID.String(), cartNullUUID.UUID.String())

	return s.cartRepository.RemoveFromCart(ctx, cartNullUUID, productNullUUID)
}

// GetCartItems returns all items in the cart
func (s *CartService) GetCartItems(ctx context.Context, cartID string) ([]model.CartItem, error) {
	// Convert the string ID to UUID
	cartUUID, err := uuid.Parse(cartID)
	if err != nil {
		s.logger.Error("failed to parse cart ID", zap.Error(err))
		return nil, fmt.Errorf("parse cart ID: %w", err)
	}

	cartNullUUID := uuid.NullUUID{UUID: cartUUID, Valid: true}

	items, err := s.cartRepository.GetCartItems(ctx, cartNullUUID)
	if err != nil {
		s.logger.Error("failed to get cart items", zap.Error(err))
		return nil, fmt.Errorf("get cart items: %w", err)
	}

	var cartItems []model.CartItem
	for _, item := range items {
		// fetch the product
		product, err := s.productRepository.GetProductByID(ctx, item.ProductID.UUID)
		if err != nil {
			s.logger.Error("failed to get product by ID", zap.Error(err))
			return nil, fmt.Errorf("get product by ID: %w", err)

		}

		// fetch the product image
		productImage, err := s.productImageRepository.ListProductImagesByProductID(ctx, uuid.NullUUID{UUID: item.ProductID.UUID, Valid: true})
		if err != nil {
			s.logger.Error("failed to get product image by product ID", zap.Error(err))
			return nil, fmt.Errorf("get product image by product ID: %w", err)
		}

		// create the cart item
		cartItem := model.CartItem{
			ID:          item.ID,
			ProductID:   item.ProductID.UUID,
			Name:        product.Name,
			Description: product.Description,
			Quantity:    item.Quantity,
			Price:       item.Price,
			ImageURL:    s.constructS3URL(productImage[0].ImageUrl),
		}

		cartItems = append(cartItems, cartItem)

	}
	return cartItems, nil
}

// ClearCart removes all items from the cart
func (s *CartService) ClearCart(ctx context.Context, cartID string) error {
	// Convert the string ID to UUID
	cartUUID, err := uuid.Parse(cartID)
	if err != nil {
		s.logger.Error("failed to parse cart ID", zap.Error(err))
		return fmt.Errorf("parse cart ID: %w", err)
	}

	cartNullUUID := uuid.NullUUID{UUID: cartUUID, Valid: true}

	return s.cartRepository.ClearCart(ctx, cartNullUUID)
}

// CalculateCartTotal calculates the total price of all items in the cart
func (s *CartService) CalculateCartTotal(ctx context.Context, cartID string) (float64, error) {
	// Convert the string ID to UUID
	cartUUID, err := uuid.Parse(cartID)
	if err != nil {
		s.logger.Error("failed to parse cart ID", zap.Error(err))
		return 0, fmt.Errorf("parse cart ID: %w", err)
	}

	cartNullUUID := uuid.NullUUID{UUID: cartUUID, Valid: true}

	return s.cartRepository.CalculateCartTotal(ctx, cartNullUUID)
}

// UpdateCartItemQuantity updates the quantity of an item in the cart
func (s *CartService) UpdateCartItemQuantity(ctx context.Context, cartID, productID string, quantity int32) error {
	// Convert the string IDs to UUID
	cartUUID, err := uuid.Parse(cartID)
	if err != nil {
		s.logger.Error("failed to parse cart ID", zap.Error(err))
		return fmt.Errorf("parse cart ID: %w", err)
	}

	productUUID, err := uuid.Parse(productID)
	if err != nil {
		s.logger.Error("failed to parse product ID", zap.Error(err))
		return fmt.Errorf("parse product ID: %w", err)
	}

	cartNullUUID := uuid.NullUUID{UUID: cartUUID, Valid: true}
	productNullUUID := uuid.NullUUID{UUID: productUUID, Valid: true}

	return s.cartRepository.UpdateCartItemQuantity(ctx, cartNullUUID, productNullUUID, quantity)
}

// CreateShoppingCart creates a new shopping cart
func (s *CartService) CreateShoppingCart(ctx context.Context, userID string) (model.ShoppingCart, error) {
	// Convert the string IDs to UUID
	var userNullUUID uuid.NullUUID

	if userID == "" {
		userNullUUID = uuid.NullUUID{UUID: uuid.Nil, Valid: false}
	} else {
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			s.logger.Error("failed to parse user ID", zap.Error(err))
			return model.ShoppingCart{}, fmt.Errorf("parse user ID: %w", err)
		}
		userNullUUID = uuid.NullUUID{UUID: userUUID, Valid: true}
	}

	cart, err := s.cartRepository.CreateShoppingCart(ctx, userNullUUID)
	if err != nil {
		s.logger.Error("failed to create shopping cart", zap.Error(err))
		return model.ShoppingCart{}, fmt.Errorf("create shopping cart: %w", err)
	}

	log.Printf("Created shopping cart with ID: %s", cart.ID.String())

	return cart, nil
}

// GetShoppingCartByUserID returns the shopping cart of a user
func (s *CartService) GetShoppingCartByUserID(ctx context.Context, userID string) (string, error) {
	// Convert the string ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Error("failed to parse user ID", zap.Error(err))
		return "", fmt.Errorf("parse user ID: %w", err)
	}

	userNullUUID := uuid.NullUUID{UUID: userUUID, Valid: true}

	cart, err := s.cartRepository.GetShoppingCartByUserID(ctx, userNullUUID)
	if err != nil {
		s.logger.Error("failed to get shopping cart by user ID", zap.Error(err))
		return "", fmt.Errorf("get shopping cart by user ID: %w", err)
	}

	return cart.SessionID.String(), nil
}

// GetShoppingCartBySessionID returns the shopping cart by session ID
func (s *CartService) GetShoppingCartBySessionID(ctx context.Context, sessionID string) (string, error) {
	// Convert the string ID to UUID
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		s.logger.Error("failed to parse session ID", zap.Error(err))
		return "", fmt.Errorf("parse session ID: %w", err)
	}

	sessionNullUUID := uuid.NullUUID{UUID: sessionUUID, Valid: true}

	cart, err := s.cartRepository.GetShoppingCartBySessionID(ctx, sessionNullUUID)
	if err != nil {
		s.logger.Error("failed to get shopping cart by session ID", zap.Error(err))
		return "", fmt.Errorf("get shopping cart by session ID: %w", err)
	}

	return cart.UserID.String(), nil
}

// DeleteShoppingCart deletes the shopping cart
func (s *CartService) DeleteShoppingCart(ctx context.Context, cartID string) error {
	// Convert the string ID to UUID
	cartUUID, err := uuid.Parse(cartID)
	if err != nil {
		s.logger.Error("failed to parse cart ID", zap.Error(err))
		return fmt.Errorf("parse cart ID: %w", err)
	}

	return s.cartRepository.DeleteShoppingCart(ctx, cartUUID)
}

// constructS3URL constructs the S3 URL for a given file path
func (s *CartService) constructS3URL(filePath string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.AWSBucketName, s.config.AWSRegion, filePath)
}
