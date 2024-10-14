// handlers/cart_handler.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"weblineBackend/internal/middleware"
	"weblineBackend/internal/services"

	"go.uber.org/zap"
)

// CartHandler handles cart-related HTTP requests.
type CartHandler struct {
	cartService *services.CartService
	logger      *zap.Logger
}

// NewCartHandler creates a new CartHandler with the given CartService and logger.
func NewCartHandler(cartService *services.CartService, logger *zap.Logger) *CartHandler {
	return &CartHandler{
		cartService: cartService,
		logger:      logger,
	}
}

// AddToCartHandler adds an item to the cart or updates the quantity if it already exists.
func (h *CartHandler) AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromContext(r)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Parse the request body.
	var req struct {
		ProductID string `json:"productID"`
		Quantity  int32  `json:"quantity"`
		Price     string `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to parse request body", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	price, err := strconv.ParseFloat(req.Price, 64)
	if err != nil {
		h.logger.Warn("Failed to parse price", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, "Invalid price")
		return
	}

	// Validate input.
	if req.ProductID == "" || req.Quantity <= 0 || price < 0 {
		RespondWithError(w, http.StatusBadRequest, "Invalid input parameters")
		return
	}

	// Call the service to add/update the cart item.
	if err := h.cartService.AddToCart(r.Context(), user.User, req.ProductID, req.Quantity, price); err != nil {
		h.logger.Error("Failed to add item to cart", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to add item to cart")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Item added to cart")
}

// RemoveFromCartHandler removes an item from the cart.
func (h *CartHandler) RemoveFromCartHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromContext(r)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Parse the request body.
	var req struct {
		ProductID string `json:"productID"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to parse request body", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input.
	if req.ProductID == "" {
		RespondWithError(w, http.StatusBadRequest, "ProductID is required")
		return
	}

	// Call the service to remove the cart item.
	if err := h.cartService.RemoveFromCart(r.Context(), user.User, req.ProductID); err != nil {
		h.logger.Error("Failed to remove item from cart", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to remove item from cart")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Item removed from cart")
}

// GetCartItemsHandler returns the items in the cart.
func (h *CartHandler) GetCartItemsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromContext(r)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Call the service to get cart items.
	items, err := h.cartService.GetCartItems(r.Context(), user.User)
	if err != nil {
		h.logger.Error("Failed to get cart items", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to get cart items")
		return
	}

	RespondWithJSON(w, http.StatusOK, items)
}

// ClearCartHandler clears the cart.
func (h *CartHandler) ClearCartHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromContext(r)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Call the service to clear the cart.
	if err := h.cartService.ClearCart(r.Context(), user.User); err != nil {
		h.logger.Error("Failed to clear cart", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to clear cart")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Cart cleared")
}

// UpdateCartItemQuantityHandler updates the quantity of an item in the cart.
func (h *CartHandler) UpdateCartItemQuantityHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromContext(r)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Parse the request body.
	var req struct {
		ProductID string `json:"productID"`
		Quantity  int32  `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to parse request body", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input.
	if req.ProductID == "" || req.Quantity < 0 {
		RespondWithError(w, http.StatusBadRequest, "Invalid input parameters")
		return
	}

	// Call the service to update the cart item quantity.
	if err := h.cartService.UpdateCartItemQuantity(r.Context(), user.User, req.ProductID, req.Quantity); err != nil {
		h.logger.Error("Failed to update item quantity", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to update item quantity")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Item quantity updated")
}

// CalculateCartTotalHandler calculates the total price of the items in the cart.
func (h *CartHandler) CalculateCartTotalHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromContext(r)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Call the service to calculate the total.
	total, err := h.cartService.CalculateCartTotal(r.Context(), user.User)
	if err != nil {
		h.logger.Error("Failed to calculate cart total", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to calculate cart total")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]float64{"total": total})
}

// ReplaceCartItemsHandler replaces the items in the cart.
func (h *CartHandler) ReplaceCartItemsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromContext(r)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Parse the request body.
	var req struct {
		Items []struct {
			ProductID string  `json:"productID"`
			Quantity  int32   `json:"quantity"`
			Price     float64 `json:"price"`
		} `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to parse request body", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input.
	if len(req.Items) == 0 {
		RespondWithError(w, http.StatusBadRequest, "No items provided")
		return
	}

	// Convert request items to service format.
	serviceItems := make([]services.CartItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		if item.ProductID == "" || item.Quantity <= 0 || item.Price < 0 {
			RespondWithError(w, http.StatusBadRequest, "Invalid item parameters")
			return
		}
		serviceItems = append(serviceItems, services.CartItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	// Call the service to replace cart items.
	if err := h.cartService.ReplaceCartItems(r.Context(), user.User, serviceItems); err != nil {
		h.logger.Error("Failed to replace cart items", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to replace cart items")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Cart items replaced")
}

// GetShoppingCartHandler returns the shopping cart for the current user.
func (h *CartHandler) GetShoppingCartHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromContext(r)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Call the service to get the shopping cart.
	cart, err := h.cartService.GetShoppingCart(r.Context(), user.User)
	if err != nil {
		h.logger.Error("Failed to get shopping cart", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to get shopping cart")
		return
	}

	RespondWithJSON(w, http.StatusOK, cart)
}

// Helper function to retrieve user from context and get cart owner ID.
func (h *CartHandler) getUserFromContext(r *http.Request) (*UserContext, error) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}

	var cartOwnerID string
	if user.IsGuest {
		cartOwnerID = user.GuestID
	} else {
		cartOwnerID = user.UserID.String()
	}

	return &UserContext{
		User:        user,
		CartOwnerID: cartOwnerID,
	}, nil
}

// UserContext holds user information and cart owner ID.
type UserContext struct {
	User        middleware.User
	CartOwnerID string
}
