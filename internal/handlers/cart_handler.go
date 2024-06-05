package handlers

import (
	"encoding/json"
	"net/http"
	"weblineBackend/internal/services"

	"github.com/gorilla/mux"
)

type CartHandler struct {
	cartService *services.CartService
}

func NewCartHandler(cartService *services.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

// AddToCartHandler adds an item to the cart or updates the quantity if it already exists
func (h *CartHandler) AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req struct {
		ProductID string  `json:"product_id"`
		Quantity  int32   `json:"quantity"`
		Price     float64 `json:"price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "failed to parse request body")
		return
	}

	// Get the cart ID from the URL
	cartID := mux.Vars(r)["cart_id"]

	// Call the service
	if err := h.cartService.AddToCart(r.Context(), cartID, req.ProductID, req.Quantity, req.Price); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to add item to cart")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "item added to cart")
}

// RemoveFromCartHandler removes an item from the cart
func (h *CartHandler) RemoveFromCartHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req struct {
		ProductID string `json:"product_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "failed to parse request body")
		return
	}

	cartID := mux.Vars(r)["cart_id"]

	// Call the service
	if err := h.cartService.RemoveFromCart(r.Context(), cartID, req.ProductID); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to remove item from cart")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "item removed from cart")
}

// GetCartItemsHandler returns the items in the cart
func (h *CartHandler) GetCartItemsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	cartID := mux.Vars(r)["cart_id"]

	// Call the service
	items, err := h.cartService.GetCartItems(r.Context(), cartID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to get cart items")
		return
	}

	// Write the response
	if err := json.NewEncoder(w).Encode(items); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to encode response")
		return
	}

	RespondWithJSON(w, http.StatusOK, items)
}

// ClearCartHandler clears the cart
func (h *CartHandler) ClearCartHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	cartID := mux.Vars(r)["cart_id"]

	// Call the service
	if err := h.cartService.ClearCart(r.Context(), cartID); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to clear cart")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "cart cleared")
}

// CalculateCartTotalHandler calculates the total price of the items in the cart
func (h *CartHandler) CalculateCartTotalHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	cartID := mux.Vars(r)["cart_id"]

	// Call the service
	total, err := h.cartService.CalculateCartTotal(r.Context(), cartID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to calculate cart total")
		return
	}

	// Write the response
	if err := json.NewEncoder(w).Encode(total); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to encode response")
		return
	}

	RespondWithJSON(w, http.StatusOK, total)
}

// UpdateCartItemQuantityHandler updates the quantity of an item in the cart
func (h *CartHandler) UpdateCartItemQuantityHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req struct {
		ProductID string `json:"product_id"`
		Quantity  int32  `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "failed to parse request body")
		return
	}

	cartID := mux.Vars(r)["cart_id"]

	// Call the service
	if err := h.cartService.UpdateCartItemQuantity(r.Context(), cartID, req.ProductID, req.Quantity); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to update item quantity")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "item quantity updated")
}

// CreateShoppingCartHandler creates a new shopping cart
func (h *CartHandler) CreateShoppingCartHandler(w http.ResponseWriter, r *http.Request) {

	var req struct {
		UserID string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "failed to parse request body")
		return
	}

	// Call the service
	cartID, err := h.cartService.CreateShoppingCart(r.Context(), req.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to create shopping cart")
		return
	}

	// Write the response
	if err := json.NewEncoder(w).Encode(cartID); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to encode response")
		return
	}

	RespondWithJSON(w, http.StatusCreated, cartID)
}

// GetShoppingCartByUserIDHandler returns the shopping cart by user ID
func (h *CartHandler) GetShoppingCartByUserIDHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	userID := r.URL.Query().Get("user_id")

	// Call the service
	cart, err := h.cartService.GetShoppingCartByUserID(r.Context(), userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to get shopping cart")
		return
	}

	// Write the response
	if err := json.NewEncoder(w).Encode(cart); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to encode response")
		return
	}

	RespondWithJSON(w, http.StatusOK, cart)
}

// GetShoppingCartBySessionIDHandler returns the shopping cart by session ID
func (h *CartHandler) GetShoppingCartBySessionIDHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	sessionID := r.URL.Query().Get("session_id")

	// Call the service
	cart, err := h.cartService.GetShoppingCartBySessionID(r.Context(), sessionID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to get shopping cart")
		return
	}

	// Write the response
	if err := json.NewEncoder(w).Encode(cart); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to encode response")
		return
	}

	RespondWithJSON(w, http.StatusOK, cart)
}

// DeleteShoppingCartHandler deletes the shopping cart
func (h *CartHandler) DeleteShoppingCartHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	sessionID := r.URL.Query().Get("session_id")

	// Call the service
	if err := h.cartService.DeleteShoppingCart(r.Context(), sessionID); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to delete shopping cart")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "shopping cart deleted")
}
