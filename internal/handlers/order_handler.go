package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"weblineBackend/internal/model"
	"weblineBackend/internal/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type OrderHandler struct {
	orderService   *services.OrderService
	paymentService *services.PaymentService
}

func NewOrderHandler(orderService *services.OrderService, paymentService *services.PaymentService) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		paymentService: paymentService,
	}
}

// CreateOrder creates a new order
type CreateOrderRequest struct {
	FirstName        string                  `json:"first_name"`
	LastName         string                  `json:"last_name"`
	StreetAddress    string                  `json:"street_address"`
	City             string                  `json:"city"`
	State            string                  `json:"state"`
	Country          string                  `json:"country"`
	Phone            string                  `json:"phone"`
	Email            string                  `json:"email"`
	ShippingOption   string                  `json:"shipping_option"`
	OrderItems       []CreateOrderItemParams `json:"order_items"`
	Total            float64                 `json:"total"`
	PaymentMethod    int32                   `json:"payment_method"`
	ShippingLocation string                  `json:"shipping_location"`
	ShippingCost     float64                 `json:"shipping_cost"`
}

type CreateOrderItemParams struct {
	ProductID        string   `json:"product_id"`
	ProductOptionIDs []string `json:"product_option_id"`
	ColorID          string   `json:"color_id"`
	SizeID           string   `json:"size_id"`
	Quantity         int32    `json:"quantity"`
	Price            float64  `json:"price"`
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Create order Params
	orderParams := &model.CreateOrderParams{
		GuestCheckoutID:  uuid.NullUUID{UUID: uuid.Nil, Valid: false},
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		StreetAddress:    req.StreetAddress,
		City:             req.City,
		State:            req.State,
		Country:          req.Country,
		Phone:            req.Phone,
		Email:            req.Email,
		ShippingOption:   req.ShippingOption,
		Total:            req.Total,
		PaymentMethod:    req.PaymentMethod,
		ShippingLocation: req.ShippingLocation,
		ShippingCost:     req.ShippingCost,
	}

	// Create order items
	var orderItems []model.CreateOrderItemParams
	for _, item := range req.OrderItems {
		if item.ProductID == "" {
			RespondWithError(w, http.StatusBadRequest, "Product ID is required")
			return
		}

		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid product ID")
			return
		}

		var productOptionIDs []uuid.NullUUID
		for _, optionID := range item.ProductOptionIDs {
			if optionID == "" {
				productOptionIDs = append(productOptionIDs, uuid.NullUUID{UUID: uuid.Nil, Valid: false})
				continue
			}

			optionUUID, err := uuid.Parse(optionID)
			if err != nil {
				RespondWithError(w, http.StatusBadRequest, "Invalid product option ID")
				return
			}

			productOptionIDs = append(productOptionIDs, uuid.NullUUID{UUID: optionUUID, Valid: true})
		}

		var colorID *uuid.NullUUID
		if item.ColorID == "" {
			colorID = &uuid.NullUUID{UUID: uuid.Nil, Valid: false}
		} else {
			colorUUID, err := uuid.Parse(item.ColorID)
			if err != nil {
				RespondWithError(w, http.StatusBadRequest, "Invalid color ID")
				return
			}
			colorID = &uuid.NullUUID{UUID: colorUUID, Valid: true}
		}

		var sizeID *uuid.NullUUID
		if item.SizeID == "" {
			sizeID = &uuid.NullUUID{UUID: uuid.Nil, Valid: false}
		} else {
			sizeUUID, err := uuid.Parse(item.SizeID)
			if err != nil {
				RespondWithError(w, http.StatusBadRequest, "Invalid size ID")
				return
			}
			sizeID = &uuid.NullUUID{UUID: sizeUUID, Valid: true}
		}

		orderItems = append(orderItems, model.CreateOrderItemParams{
			ProductID:        productID,
			ProductOptionIDs: productOptionIDs,
			ColorID:          *colorID,
			SizeID:           *sizeID,
			Quantity:         item.Quantity,
			Price:            strconv.FormatFloat(item.Price, 'f', -1, 64),
		})
	}

	// Create order
	orderID, err := h.orderService.CreateOrder(r.Context(), orderParams, orderItems)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write response
	RespondWithJSON(w, http.StatusCreated, orderID)
}

// ListOrders lists all orders for authenticated user
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value("userId").(uuid.UUID)

	if userID == uuid.Nil {
		log.Println("Unauthorized")
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get orders
	orders, err := h.orderService.ListOrders(r.Context(), userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list orders: %v", err))
		return
	}

	// Write response
	RespondWithJSON(w, http.StatusOK, orders)
}

// GetOrder gets a single order
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	// Get order ID from URL
	orderID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	// Get order
	order, err := h.orderService.GetOrder(r.Context(), orderID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get order: %v", err))
		return
	}

	// Write response
	RespondWithJSON(w, http.StatusOK, order)
}

type PayOrderRequest struct {
	OrderID     string `json:"orderID"`
	PhoneNumber string `json:"phoneNumber"`
}

// PayOrder pays for an order
func (h *OrderHandler) PayOrder(w http.ResponseWriter, r *http.Request) {
	// Get payment method from query
	method := r.URL.Query().Get("method")

	if method == "" {
		RespondWithError(w, http.StatusBadRequest, "Payment method is required")
		return
	}

	// Get payment method from request
	var req PayOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	switch method {
	case "mpesa":
		// Pay with M-Pesa
		response, err := h.paymentService.PayOrderWithMpesa(r.Context(), req.OrderID, req.PhoneNumber)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to pay order: %v", err))
			return
		}

		// Write response
		RespondWithJSON(w, http.StatusOK, response)

	default:
		RespondWithError(w, http.StatusBadRequest, "Invalid payment method")
	}
}
