package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"weblineBackend/internal/middleware"
	"weblineBackend/internal/model"
	"weblineBackend/internal/services"
	"weblineBackend/pkg/mpesa"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type OrderHandler struct {
	logger         *zap.Logger
	orderService   *services.OrderService
	paymentService *services.PaymentService
	userService    *services.UserService
}

func NewOrderHandler(logger *zap.Logger, orderService *services.OrderService, paymentService *services.PaymentService, userService *services.UserService) *OrderHandler {
	return &OrderHandler{
		logger:         logger,
		orderService:   orderService,
		paymentService: paymentService,
		userService:    userService,
	}
}

// CreateOrderRequest creates a new order
type CreateOrderRequest struct {
	FirstName        string                  `json:"firstName"`
	LastName         string                  `json:"lastName"`
	Country          string                  `json:"country"`
	KraPIN           string                  `json:"kraPIN"`
	CompanyName      string                  `json:"companyName"`
	City             string                  `json:"city"`
	County           string                  `json:"county"`
	Phone            string                  `json:"phone"`
	Email            string                  `json:"email"`
	CanCreateAccount bool                    `json:"canCreateAccount"`
	Password         string                  `json:"password,omitempty"`
	OrderNotes       string                  `json:"orderNotes"`
	OrderItems       []CreateOrderItemParams `json:"orderItems"`
}

type CreateOrderItemParams struct {
	ProductID        string   `json:"productID"`
	ProductOptionIDs []string `json:"productOptionIDs"`
	ColorID          string   `json:"colorID"`
	SizeID           string   `json:"sizeID"`
	Quantity         int32    `json:"quantity"`
	Price            float64  `json:"price"`
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validateCreateOrderRequest(&req); err != nil {
		h.logger.Error("Invalid create order request", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Retrieve userID from context
	userID, isAuthenticated := middleware.GetUserID(r.Context())

	var userUUID *uuid.UUID
	var guestUUID *uuid.UUID
	var err error

	if isAuthenticated {
		// User is logged in
		userUUID = &userID
		err = h.userService.UpdateUserInfo(r.Context(), model.UpdateUserInfoParams{
			FirstName:   req.FirstName,
			LastName:    req.LastName,
			PhoneNumber: req.Phone,
		}, nil)
		if err != nil {
			h.logger.Error("Failed to update user profile", zap.Error(err))
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user profile")
			return
		}
	} else {
		// User is not logged in
		userUUID, guestUUID, err = h.handleGuestOrExistingUser(r.Context(), &req)
		if err != nil {
			h.logger.Error("Failed to handle guest or existing user", zap.Error(err))
			RespondWithError(w, http.StatusInternalServerError, "Failed to process user information")
			return
		}
	}

	orderParams := &model.CreateOrderParams{
		GuestID:     guestUUID,
		UserID:      userUUID,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Country:     req.Country,
		KraPIN:      &req.KraPIN,
		CompanyName: &req.CompanyName,
		City:        req.City,
		County:      req.County,
		Phone:       req.Phone,
		Email:       req.Email,
		OrderNotes:       &req.OrderNotes,
		CanCreateAccount: req.CanCreateAccount,
		Password:         &req.Password,
	}

	// Convert OrderItems directly from request
	orderItems, err := h.createOrderItems(req.OrderItems)
	if err != nil {
		h.logger.Error("Failed to create order items", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to create order items")
		return
	}

	orderID, err := h.orderService.CreateOrder(r.Context(), orderParams, orderItems)
	if err != nil {
		h.logger.Error("Failed to create order", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to create order")
		return
	}

	RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"orderID": orderID,
	})
}

func (h *OrderHandler) handleGuestOrExistingUser(ctx context.Context, req *CreateOrderRequest) (*uuid.UUID, *uuid.UUID, error) {
	// Check if the email already exists
	existingUser, err := h.userService.GetUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}

	if existingUser != nil {
		// Email exists, return the user ID
		return &existingUser.ID, nil, nil
	}

	if req.CanCreateAccount {
		// User chose to create an account
		newUserID, err := h.userService.CreateUserFromOrder(ctx, model.CreateUserParams{
			Email:       req.Email,
			Password:    req.Password,
			FirstName:   req.FirstName,
			LastName:    req.LastName,
			PhoneNumber: req.Phone,
		})
		if err != nil {
			return nil, nil, err
		}
		return &newUserID, nil, nil
	}

	// Create guest user
	guestID, err := h.userService.CreateGuestUser(ctx, model.CreateGuestUserParams{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		City:      req.City,
		County:    req.County,
		Country:   req.Country,
	})
	if err != nil {
		return nil, nil, err
	}

	return nil, guestID, nil
}

func (h *OrderHandler) validateCreateOrderRequest(req *CreateOrderRequest) error {
	if req.FirstName == "" || req.LastName == "" || req.Country == "" || req.City == "" || req.County == "" || req.Phone == "" || req.Email == "" {
		h.logger.Error("missing required fields", zap.Any("request", req))
		return errors.New("missing required fields")
	}

	if !slices.Contains(model.AVAILABLE_COUNTRIES, req.Country) {
		h.logger.Error("invalid country", zap.String("country", req.Country))
		return errors.New("invalid country")
	}

	if !slices.Contains(model.COUNTIES, req.County) {
		h.logger.Error("invalid county", zap.String("county", req.County))
		return errors.New("invalid county")
	}

	if req.CanCreateAccount && req.Password == "" {
		h.logger.Error("password is required when creating an account", zap.Any("request", req))
		return errors.New("password is required when creating an account")
	}

	if len(req.OrderItems) == 0 {
		h.logger.Error("order must contain at least one item", zap.Any("request", req))
		return errors.New("order must contain at least one item")
	}

	return nil
}

func (h *OrderHandler) createOrderItems(items []CreateOrderItemParams) ([]model.CreateOrderItemParams, error) {
	var orderItems []model.CreateOrderItemParams
	for _, item := range items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID: %s", item.ProductID)
		}

		productOptionIDs := make([]uuid.NullUUID, len(item.ProductOptionIDs))
		for i, optionID := range item.ProductOptionIDs {
			if optionID == "" {
				productOptionIDs[i] = uuid.NullUUID{UUID: uuid.Nil, Valid: false}
			} else {
				optionUUID, err := uuid.Parse(optionID)
				if err != nil {
					return nil, fmt.Errorf("invalid product option ID: %s", optionID)
				}
				productOptionIDs[i] = uuid.NullUUID{UUID: optionUUID, Valid: true}
			}
		}

		colorID := h.parseOptionalUUID(item.ColorID)
		sizeID := h.parseOptionalUUID(item.SizeID)

		orderItems = append(orderItems, model.CreateOrderItemParams{
			ProductID:        productID,
			ProductOptionIDs: productOptionIDs,
			ColorID:          colorID,
			SizeID:           sizeID,
			Quantity:         item.Quantity,
			Price:            strconv.FormatFloat(item.Price, 'f', -1, 64),
		})
	}
	return orderItems, nil
}

func (h *OrderHandler) parseOptionalUUID(id string) uuid.NullUUID {
	if id == "" {
		return uuid.NullUUID{UUID: uuid.Nil, Valid: false}
	}
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return uuid.NullUUID{UUID: uuid.Nil, Valid: false}
	}
	return uuid.NullUUID{UUID: parsedID, Valid: true}
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userId").(uuid.UUID)
	if !ok || userID == uuid.Nil {
		h.logger.Error("Unauthorized access attempt")
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orders, err := h.orderService.ListOrders(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to list orders", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to list orders")
		return
	}

	RespondWithJSON(w, http.StatusOK, orders)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		h.logger.Error("Invalid order ID", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	order, err := h.orderService.GetOrder(r.Context(), orderID)
	if err != nil {
		h.logger.Error("Failed to get order", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to get order")
		return
	}

	RespondWithJSON(w, http.StatusOK, order)
}

type PayOrderRequest struct {
	OrderID     string `json:"orderID"`
	PhoneNumber string `json:"phoneNumber"`
}

func (h *OrderHandler) PayOrder(w http.ResponseWriter, r *http.Request) {
	method := r.URL.Query().Get("method")
	if method == "" {
		RespondWithError(w, http.StatusBadRequest, "Payment method is required")
		return
	}

	var req PayOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request payload", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	switch method {
	case "mpesa":
		err := h.paymentService.PayOrderWithMpesa(r.Context(), req.OrderID, req.PhoneNumber)
		if err != nil {
			h.logger.Error("Failed to pay order with M-Pesa", zap.Error(err))
			RespondWithError(w, http.StatusInternalServerError, "Failed to process payment")
			return
		}
		RespondWithSuccess(w, http.StatusOK, "Order paid successfully")
	default:
		RespondWithError(w, http.StatusBadRequest, "Invalid payment method")
	}
}

func (h *OrderHandler) HandleMpesaCallback(w http.ResponseWriter, r *http.Request) {
	var callbackResponse mpesa.MpesaCallbackResponse

	h.logger.Info("Processing Mpesa callback")
	if err := json.NewDecoder(r.Body).Decode(&callbackResponse); err != nil {
		h.logger.Error("Failed to decode callback response", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, "Invalid callback payload")
		return
	}

	err := h.paymentService.ProcessMpesaCallback(r.Context(), callbackResponse)
	if err != nil {
		h.logger.Error("Failed to process Mpesa callback", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to process callback")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Callback processed successfully"})
}

func (h *OrderHandler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("orderID")
	if orderID == "" {
		h.logger.Error("Order ID is required")
		RespondWithError(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	status, err := h.paymentService.GetPaymentStatus(r.Context(), orderID)
	if err != nil {
		h.logger.Error("Failed to get payment status", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to get payment status")
		return
	}

	RespondWithJSON(w, http.StatusOK, status)
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	orderID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		h.logger.Error("Invalid order ID", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request payload", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = h.orderService.CancelOrder(r.Context(), orderID, req.Reason)
	if err != nil {
		h.logger.Error("Failed to cancel order", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to cancel order")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Order cancelled successfully")
}

func (h *OrderHandler) ChangeOrderPaymentMethod(w http.ResponseWriter, r *http.Request) {
	orderID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		h.logger.Error("Invalid order ID", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var req struct {
		Method string `json:"method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request payload", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = h.orderService.ChangeOrderPaymentMethod(r.Context(), orderID, req.Method)
	if err != nil {
		h.logger.Error("Failed to change payment method", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to change payment method")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Payment method changed successfully")
}

func (h *OrderHandler) GetTotalRevenue(w http.ResponseWriter, r *http.Request) {
	totalRevenue, err := h.orderService.GetTotalRevenue(r.Context())
	if err != nil {
		h.logger.Error("Failed to get total revenue", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to get total revenue")
		return
	}

	RespondWithJSON(w, http.StatusOK, totalRevenue)
}

func (h *OrderHandler) GetMonthlySales(w http.ResponseWriter, r *http.Request) {
	revenue, err := h.orderService.GetMonthlySales(r.Context())
	if err != nil {
		h.logger.Error("Failed to get monthly sales", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to get monthly sales")
		return
	}

	RespondWithJSON(w, http.StatusOK, revenue)
}

func (h *OrderHandler) GetMonthlyRevenue(w http.ResponseWriter, r *http.Request) {
	revenue, err := h.orderService.GetMonthlyRevenue(r.Context())
	if err != nil {
		h.logger.Error("Failed to get monthly revenue", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to get monthly revenue")
		return
	}

	RespondWithJSON(w, http.StatusOK, revenue)
}

func (h *OrderHandler) GetSalesTrend(w http.ResponseWriter, r *http.Request) {
	revenue, err := h.orderService.GetSalesTrend(r.Context())
	if err != nil {
		h.logger.Error("Failed to get sales trend", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to get sales trend")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]float64{"revenue": revenue})
}

// GetRecentSales returns the total revenue for the last two months
func (h *OrderHandler) GetRecentSales(w http.ResponseWriter, r *http.Request) {
	revenue, err := h.orderService.GetRecentSales(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get recent sales: %v", err))
		return
	}

	RespondWithJSON(w, http.StatusOK, revenue)
}

// GetTotalSalesCurrentMonth returns the total revenue for the current month
func (h *OrderHandler) GetTotalSalesCurrentMonth(w http.ResponseWriter, r *http.Request) {
	revenue, err := h.orderService.GetTotalSalesCurrentMonth(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get total sales for current month: %v", err))
		return
	}

	RespondWithJSON(w, http.StatusOK, revenue)
}

// GetExchangeRate returns the exchange rate
func (h *OrderHandler) GetExchangeRate(w http.ResponseWriter, r *http.Request) {
	rate, err := h.orderService.GetExchangeRate(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get exchange rate: %v", err))
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]float64{"rate": rate})
}

// UpdateExchangeRate updates the exchange rate
func (h *OrderHandler) UpdateExchangeRate(w http.ResponseWriter, r *http.Request) {
	// Get rate from body
	var req struct {
		Rate float64 `json:"rate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Rate <= 0 {
		RespondWithError(w, http.StatusBadRequest, "Invalid exchange rate")
		return
	}

	// Update rate
	err := h.orderService.UpdateExchangeRate(r.Context(), req.Rate)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update exchange rate: %v", err))
		return
	}

	// Write response
	RespondWithSuccess(w, http.StatusOK, "Exchange rate updated successfully")
}
