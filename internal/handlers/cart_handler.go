// handlers/cart_handler.go
package handlers

import (
	"encoding/json"
	"net/http"
	"weblineBackend/internal/middleware"
	"weblineBackend/internal/model"
	"weblineBackend/internal/services"

	"go.uber.org/zap"
)

// CartHandler handles cart-related HTTP requests.
type CartHandler struct {
	cartService    *services.CartService
	sessionService *services.SessionService
	logger         *zap.Logger
}

// NewCartHandler creates a new CartHandler
func NewCartHandler(cartService *services.CartService, sessionService *services.SessionService, logger *zap.Logger) *CartHandler {
	return &CartHandler{
		cartService:    cartService,
		sessionService: sessionService,
		logger:         logger,
	}
}

// AddToCartHandler adds an item to the cart
func (h *CartHandler) AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	// Get session from context
	session, err := middleware.GetSessionFromContext(r.Context())
	if err != nil {
		// If the error indicates expiration, set a custom header
		if err.Error() == "session expired" {
			w.Header().Set("X-Session-Expired", "true")
		}
		h.logger.Error("Failed to get session", zap.Error(err))
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	// Update session last activity
	if err := h.sessionService.UpdateSessionLastActivity(r.Context(), session.SessionID.String()); err != nil {
		h.logger.Warn("Failed to update session last activity", zap.Error(err))
	}
	// Retrieve user type from context
	userType := middleware.GetUserType(r.Context())

	var req struct {
		ProductID string `json:"productId" validate:"required,uuid"`
		Quantity  int32  `json:"quantity" validate:"required,min=1,max=100"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to parse request body", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Call service with session context
	if err := h.cartService.AddToCart(r.Context(), session, userType, req.ProductID, req.Quantity); err != nil {
		h.logger.Error("Failed to add item to cart",
			zap.Error(err),
			zap.String("userType", userType))
		RespondWithError(w, http.StatusInternalServerError, "Failed to add item to cart")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Item added to cart")
}

// RemoveFromCartHandler removes an item from the cart.
func (h *CartHandler) RemoveFromCartHandler(w http.ResponseWriter, r *http.Request) {
	session, err := middleware.GetSessionFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get session", zap.Error(err))
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	// Retrieve user type from context
	userType := middleware.GetUserType(r.Context())

	// Parse the request body.
	var req struct {
		ProductID string `json:"productId" validate:"required,uuid"`
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
	if err := h.cartService.RemoveFromCart(r.Context(), session, userType, req.ProductID); err != nil {
		h.logger.Error("Failed to remove item from cart", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to remove item from cart")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Item removed from cart")
}

// GetCartItemsHandler returns the items in the cart.
func (h *CartHandler) GetCartItemsHandler(w http.ResponseWriter, r *http.Request) {
	session, err := middleware.GetSessionFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get session", zap.Error(err))
		RespondWithError(w, http.StatusUnauthorized, "Invalid session")
		return
	}

	// Retrieve user type from context
	userType := middleware.GetUserType(r.Context())

	// Call the service to get the cart items.
	items, err := h.cartService.GetCartItems(r.Context(), session, userType)
	if err != nil {
		h.logger.Error("Failed to get cart items",
			zap.Error(err),
			zap.String("userType", userType))
		RespondWithError(w, http.StatusInternalServerError, "Failed to get cart items")
		return
	}

	if items == nil {
		RespondWithJSON(w, http.StatusOK, []model.CartItem{})
		return
	}

	RespondWithJSON(w, http.StatusOK, items)
}

// ClearCartHandler clears the cart.
func (h *CartHandler) ClearCartHandler(w http.ResponseWriter, r *http.Request) {
	session, err := middleware.GetSessionFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get session", zap.Error(err))
		RespondWithError(w, http.StatusUnauthorized, "Invalid session")
		return
	}
	// Retrieve user type from context
	userType := middleware.GetUserType(r.Context())

	// Call the service to clear the cart.
	if err := h.cartService.ClearCart(r.Context(), session, userType); err != nil {
		h.logger.Error("Failed to clear cart", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to clear cart")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Cart cleared")
}

// UpdateCartItemQuantityHandler updates the quantity of an item in the cart.
func (h *CartHandler) UpdateCartItemQuantityHandler(w http.ResponseWriter, r *http.Request) {
	session, err := middleware.GetSessionFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get session", zap.Error(err))
		RespondWithError(w, http.StatusUnauthorized, "Invalid session")
		return
	}
	// Retrieve user type from context
	userType := middleware.GetUserType(r.Context())

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
	if err := h.cartService.UpdateCartItemQuantity(r.Context(), session, userType, req.ProductID, req.Quantity); err != nil {
		h.logger.Error("Failed to update item quantity", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to update item quantity")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Item quantity updated")
}

// CalculateCartTotalHandler calculates the total price of the items in the cart.
func (h *CartHandler) CalculateCartTotalHandler(w http.ResponseWriter, r *http.Request) {
	session, err := middleware.GetSessionFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get session", zap.Error(err))
		RespondWithError(w, http.StatusUnauthorized, "Invalid session")
		return
	}
	// Retrieve user type from context
	userType := middleware.GetUserType(r.Context())

	// Call the service to calculate the total.
	total, err := h.cartService.CalculateCartTotal(r.Context(), session, userType)
	if err != nil {
		h.logger.Error("Failed to calculate cart total", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to calculate cart total")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]float64{"total": total})
}

// ReplaceCartItemsHandler replaces the items in the cart.
func (h *CartHandler) ReplaceCartItemsHandler(w http.ResponseWriter, r *http.Request) {
	session, err := middleware.GetSessionFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get session", zap.Error(err))
		RespondWithError(w, http.StatusUnauthorized, "Invalid session")
		return
	}
	// Retrieve user type from context
	userType := middleware.GetUserType(r.Context())

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
	if err := h.cartService.ReplaceCartItems(r.Context(), session, userType, serviceItems); err != nil {
		h.logger.Error("Failed to replace cart items", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to replace cart items")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Cart items replaced")
}

// GetShoppingCartHandler returns the shopping cart for the current user.
func (h *CartHandler) GetShoppingCartHandler(w http.ResponseWriter, r *http.Request) {
	session, err := middleware.GetSessionFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get session", zap.Error(err))
		RespondWithError(w, http.StatusUnauthorized, "Invalid session")
		return
	}
	// Retrieve user type from context
	userType := middleware.GetUserType(r.Context())

	// Call the service to get the shopping cart.
	cart, err := h.cartService.GetShoppingCart(r.Context(), session, userType)
	if err != nil {
		h.logger.Error("Failed to get shopping cart", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to get shopping cart")
		return
	}

	RespondWithJSON(w, http.StatusOK, cart)
}
