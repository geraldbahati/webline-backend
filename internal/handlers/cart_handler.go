// handlers/cart_handler.go
package handlers

import (
	"encoding/json"
	"net/http"
	"weblineBackend/internal/middleware"
	"weblineBackend/internal/services"

	"github.com/gorilla/mux"
)

// CartHandler handles cart-related HTTP requests.
type CartHandler struct {
	cartService *services.CartService
}

// NewCartHandler creates a new CartHandler with the given CartService.
func NewCartHandler(cartService *services.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

// AddToCartHandler adds an item to the cart or updates the quantity if it already exists.
func (h *CartHandler) AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session from the context.
	session, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "Session not found")
		return
	}

	// Retrieve the user ID from the context, if available.
	userID, userAuthenticated := middleware.GetUserID(r.Context())

	// Parse the request body.
	var req struct {
		ProductID string  `json:"productID"`
		Quantity  int32   `json:"quantity"`
		Price     float64 `json:"price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to parse request body")
		return
	}

	// Validate input.
	if req.ProductID == "" || req.Quantity <= 0 || req.Price < 0 {
		RespondWithError(w, http.StatusBadRequest, "Invalid input parameters")
		return
	}

	// Determine the cart ID based on authentication.
	var cartID string
	var err error
	if userAuthenticated {
		// If the user is authenticated, retrieve the cart associated with the user.
		cartID, err = h.cartService.GetShoppingCartByUserID(r.Context(), userID.String())
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve user cart")
			return
		}
	} else {
		// If the user is a guest, retrieve the cart associated with the session.
		cartID, err = h.cartService.GetShoppingCartBySessionID(r.Context(), session.SessionID.String())
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve session cart")
			return
		}

	}

	// Call the service to add/update the cart item.
	if err := h.cartService.AddToCart(r.Context(), cartID, req.ProductID, req.Quantity, req.Price); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to add item to cart")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Item added to cart")
}

// RemoveFromCartHandler removes an item from the cart.
func (h *CartHandler) RemoveFromCartHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session from the context.
	session, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "Session not found")
		return
	}

	// Retrieve the user ID from the context, if available.
	userID, userAuthenticated := middleware.GetUserID(r.Context())

	// Parse the request body.
	var req struct {
		ProductID string `json:"productID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to parse request body")
		return
	}

	// Validate input.
	if req.ProductID == "" {
		RespondWithError(w, http.StatusBadRequest, "ProductID is required")
		return
	}

	// Determine the cart ID based on authentication.
	var cartID string
	var err error
	if userAuthenticated {
		// If the user is authenticated, retrieve the cart associated with the user.
		cartID, err = h.cartService.GetShoppingCartByUserID(r.Context(), userID.String())
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve user cart")
			return
		}
	} else {
		// If the user is a guest, retrieve the cart associated with the session.
		cartID, err = h.cartService.GetShoppingCartBySessionID(r.Context(), session.SessionID.String())
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve session cart")
			return
		}
	}

	// Call the service to remove the cart item.
	if err := h.cartService.RemoveFromCart(r.Context(), cartID, req.ProductID); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to remove item from cart")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Item removed from cart")
}

// GetCartItemsHandler returns the items in the cart.
func (h *CartHandler) GetCartItemsHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session from the context.
	session, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "Session not found")
		return
	}

	// Retrieve the user ID from the context, if available.
	userID, userAuthenticated := middleware.GetUserID(r.Context())

	// Determine the cart ID based on authentication.
	var cartID string
	var err error
	if userAuthenticated {
		cartID, err = h.cartService.GetShoppingCartByUserID(r.Context(), userID.String())
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve user cart")
			return
		}
	} else {
		cartID, err = h.cartService.GetShoppingCartBySessionID(r.Context(), session.SessionID.String())
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve session cart")
			return
		}
	}

	// Call the service to get cart items.
	items, err := h.cartService.GetCartItems(r.Context(), cartID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get cart items")
		return
	}

	RespondWithJSON(w, http.StatusOK, items)
}

// ClearCartHandler clears the cart.
func (h *CartHandler) ClearCartHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session from the context.
	session, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "Session not found")
		return
	}

	// Retrieve the user ID from the context, if available.
	userID, userAuthenticated := middleware.GetUserID(r.Context())

	// Determine the cart ID based on authentication.
	var cartID string
	var err error
	if userAuthenticated {
		cartID, err = h.cartService.GetShoppingCartByUserID(r.Context(), userID.String())
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve user cart")
			return
		}
	} else {
		cartID, err = h.cartService.GetShoppingCartBySessionID(r.Context(), session.SessionID.String())
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve session cart")
			return
		}
	}

	// Call the service to clear the cart.
	if err := h.cartService.ClearCart(r.Context(), cartID); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to clear cart")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Cart cleared")
}

// CalculateCartTotalHandler calculates the total price of the items in the cart.
func (h *CartHandler) CalculateCartTotalHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session from the context.
	session, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "Session not found")
		return
	}

	// Retrieve the user ID from the context, if available.
	userID, userAuthenticated := middleware.GetUserID(r.Context())

	// Determine the cart ID based on authentication.
	var cartID string
	var err error
	if userAuthenticated {
		cartID, err = h.cartService.GetShoppingCartByUserID(r.Context(), userID.String())
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve user cart")
			return
		}
	} else {
		cartID, err = h.cartService.GetShoppingCartBySessionID(r.Context(), session.SessionID.String())
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve session cart")
			return
		}
	}

	// Call the service to calculate the total.
	total, err := h.cartService.CalculateCartTotal(r.Context(), cartID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to calculate cart total")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]float64{"total": total})
}

// UpdateCartItemQuantityHandler updates the quantity of an item in the cart.
func (h *CartHandler) UpdateCartItemQuantityHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session from the context.
	session, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "Session not found")
		return
	}

	// Retrieve the user ID from the context, if available.
	userID, userAuthenticated := middleware.GetUserID(r.Context())

	// Parse the request body.
	var req struct {
		ProductID string `json:"productID"`
		Quantity  int32  `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to parse request body")
		return
	}

	// Validate input.
	if req.ProductID == "" || req.Quantity < 0 {
		RespondWithError(w, http.StatusBadRequest, "Invalid input parameters")
		return
	}

	// Determine the cart ID based on authentication.
	var cartID string
	var err error
	if userAuthenticated {
		cartID, err = h.cartService.GetShoppingCartByUserID(r.Context(), userID.String())
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve user cart")
			return
		}
	} else {
		cartID, err = h.cartService.GetShoppingCartBySessionID(r.Context(), session.SessionID.String())
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve session cart")
			return
		}
	}

	// Call the service to update the cart item quantity.
	if err := h.cartService.UpdateCartItemQuantity(r.Context(), cartID, req.ProductID, req.Quantity); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update item quantity")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Item quantity updated")
}

// CreateShoppingCartHandler creates a new shopping cart.
func (h *CartHandler) CreateShoppingCartHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session from the context.
	session, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "Session not found")
		return
	}

	// Retrieve the user ID from the context, if available.
	userID, _ := middleware.GetUserID(r.Context())

	// Call the service to create a new shopping cart.
	cart, err := h.cartService.CreateShoppingCart(r.Context(), &session.SessionID, &userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create shopping cart")
		return
	}

	RespondWithJSON(w, http.StatusCreated, cart)
}

// GetShoppingCartByUserIDHandler returns the shopping cart by user ID.
func (h *CartHandler) GetShoppingCartByUserIDHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the user ID from the context.
	userID, userAuthenticated := middleware.GetUserID(r.Context())
	if !userAuthenticated {
		RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Call the service to get the shopping cart.
	cartID, err := h.cartService.GetShoppingCartByUserID(r.Context(), userID.String())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get shopping cart")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"cartID": cartID})
}

// GetShoppingCartBySessionIDHandler returns the shopping cart by session ID.
func (h *CartHandler) GetShoppingCartBySessionIDHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session from the context.
	session, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "Session not found")
		return
	}

	// Call the service to get the shopping cart.
	userID, err := h.cartService.GetShoppingCartBySessionID(r.Context(), session.SessionID.String())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get shopping cart")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"userID": userID})
}

// // DeleteShoppingCartHandler deletes the shopping cart.
// func (h *CartHandler) DeleteShoppingCartHandler(w http.ResponseWriter, r *http.Request) {
// 	// Retrieve the session from the context.
// 	session, ok := middleware.GetSessionFromContext(r.Context())
// 	if !ok {
// 		RespondWithError(w, http.StatusUnauthorized, "Session not found")
// 		return
// 	}

// 	// Retrieve the user ID from the context, if available.
// 	userID, userAuthenticated := middleware.GetUserID(r.Context())

// 	// Determine the cart ID based on authentication.
// 	var cartID string
// 	var err error
// 	if userAuthenticated {
// 		cartID, err = h.cartService.GetShoppingCartByUserID(r.Context(), userID.String())
// 		if err != nil {
// 			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve user cart")
// 			return
// 		}
// 	} else {
// 		cartID, err = h.cartService.GetShoppingCartBySessionID(r.Context(), session.SessionID.String())
// 		if err != nil {
// 			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve session cart")
// 			return
// 		}
// 	}

// 	// Call the service to delete the shopping cart.
// 	if err := h.cartService.DeleteShoppingCart(r.Context(), cartID); err != nil {
// 		RespondWithError(w, http.StatusInternalServerError, "Failed to delete shopping cart")
// 		return
// 	}

// 	RespondWithSuccess(w, http.StatusOK, "Shopping cart deleted")
// }

// GetShoppingCartByIDHandler returns the shopping cart by ID.
func (h *CartHandler) GetShoppingCartByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the cart ID from the URL.
	cartID := mux.Vars(r)["cartID"]
	if cartID == "" {
		RespondWithError(w, http.StatusBadRequest, "CartID is required")
		return
	}

	// Call the service to get the shopping cart.
	cartID, err := h.cartService.GetShoppingCartBySessionID(r.Context(), cartID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get shopping cart")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"cartID": cartID})
}

// ReplaceCartItemsHandler replaces the items in the cart.
func (h *CartHandler) ReplaceCartItemsHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session from the context.
	session, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "Session not found")
		return
	}

	// Retrieve the user ID from the context, if available.
	userID, userAuthenticated := middleware.GetUserID(r.Context())

	// Parse the request body.
	var req struct {
		Items []struct {
			ProductID string  `json:"productID"`
			Quantity  int32   `json:"quantity"`
			Price     float64 `json:"price"`
		} `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to parse request body")
		return
	}

	// Validate input.
	if len(req.Items) == 0 {
		RespondWithError(w, http.StatusBadRequest, "No items provided")
		return
	}

	// Determine the cart ID based on authentication.
	var cartID string
	var err error
	if userAuthenticated {
		cartID, err = h.cartService.GetShoppingCartByUserID(r.Context(), userID.String())
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve user cart")
			return
		}
	} else {
		cartID, err = h.cartService.GetShoppingCartBySessionID(r.Context(), session.SessionID.String())
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve session cart")
			return
		}
	}

	// Convert request items to service format.
	var serviceItems []services.CartItemInput
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
	if err := h.cartService.ReplaceCartItems(r.Context(), cartID, serviceItems); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to replace cart items")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Cart items replaced")
}
