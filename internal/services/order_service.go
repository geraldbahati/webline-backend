package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
	"weblineBackend/pkg/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrderService struct {
	logger              *zap.Logger
	guestCheckoutRepo   *repository.GuestCheckoutRepository
	orderRepository     *repository.OrderRepository
	orderItemRepository *repository.OrderItemRepository
	paymentRepository   *repository.PaymentRepository
	userRepo            *repository.UserRepository
	productRepo         *repository.ProductRepository
	cfg                 *appconfig.Config
}

func NewOrderService(logger *zap.Logger, guestCheckoutRepo *repository.GuestCheckoutRepository, orderRepository *repository.OrderRepository, orderItemRepository *repository.OrderItemRepository, paymentRepository *repository.PaymentRepository, userRepo *repository.UserRepository, productRepo *repository.ProductRepository, cfg *appconfig.Config) *OrderService {
	return &OrderService{
		logger:              logger,
		guestCheckoutRepo:   guestCheckoutRepo,
		orderRepository:     orderRepository,
		orderItemRepository: orderItemRepository,
		paymentRepository:   paymentRepository,
		userRepo:            userRepo,
		productRepo:         productRepo,
		cfg:                 cfg,
	}
}

type CreateGuestCheckoutParams struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Phone         string    `json:"phone"`
	StreetAddress string    `json:"street_address"`
	City          string    `json:"city"`
	State         string    `json:"state"`
	Country       string    `json:"country"`
}

type CreateOrderParams struct {
	ID              uuid.UUID  `json:"id"`
	UserID          *uuid.UUID `json:"user_id,omitempty"`
	GuestCheckoutID *uuid.UUID `json:"guest_checkout_id,omitempty"`
	Total           float64    `json:"total"`
}

type OrderResponse struct {
	OrderID   uuid.UUID `json:"orderID"`
	PayingNow bool      `json:"payingNow"`
}

// CreateOrder creates a new order with items
func (s *OrderService) CreateOrder(ctx context.Context, orderParams *model.CreateOrderParams, items []model.CreateOrderItemParams) (*OrderResponse, error) {
	userID := ctx.Value("userId")
	var userUUID uuid.NullUUID

	// Create or retrieve user or guest checkout
	if userID != nil {
		user, err := s.getUserUUID(ctx, userID.(uuid.UUID))
		if err != nil {
			return nil, err
		}

		userUUID = user
	} else {
		guestUUID, err := s.createOrRetrieveGuest(ctx, orderParams)
		if err != nil {
			return nil, err
		}
		userUUID = uuid.NullUUID{UUID: uuid.Nil, Valid: false}
		orderParams.GuestCheckoutID = uuid.NullUUID{UUID: guestUUID, Valid: true}
	}

	orderID, err := s.createOrderRecord(ctx, orderParams, userUUID)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if err := s.createOrderItem(ctx, orderID, item); err != nil {
			return nil, err
		}
	}

	orderPayment, err := s.createPayment(ctx, orderID, orderParams.Total, orderParams.PaymentOption)
	if err != nil {
		return nil, err
	}

	orderItems := make([]utils.OrderItem, 0)
	for _, item := range items {
		// get the product by id
		product, err := s.productRepo.GetProductByID(ctx, item.ProductID)
		if err != nil {
			s.logger.Error("failed to get product", zap.Error(err))
			return nil, fmt.Errorf("failed to get product: %w", err)
		}

		price, err := strconv.ParseFloat(product.Price, 64)
		if err != nil {
			s.logger.Error("failed to parse price", zap.Error(err))
			return nil, fmt.Errorf("failed to parse price: %w", err)
		}

		orderItems = append(orderItems, utils.OrderItem{
			ProductName: product.Name,
			Quantity:    item.Quantity,
			Price:       price,
		})
	}

	payingNow := orderParams.PaymentOption == "now"

	if err := utils.SendOrderNotification(s.cfg, orderID, orderParams, orderItems); err != nil {
		s.logger.Error("failed to send order notification", zap.Error(err))
	}

	return &OrderResponse{OrderID: orderPayment.OrderID, PayingNow: payingNow}, nil
}

func (s *OrderService) getUserUUID(ctx context.Context, userID uuid.UUID) (uuid.NullUUID, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user", zap.Error(err))
		return uuid.NullUUID{}, fmt.Errorf("failed to get user: %w", err)
	}
	return uuid.NullUUID{UUID: user.ID, Valid: true}, nil
}

func (s *OrderService) createOrRetrieveGuest(ctx context.Context, orderParams *model.CreateOrderParams) (uuid.UUID, error) {
	existingGuest, err := s.guestCheckoutRepo.GetGuestCheckoutByEmail(ctx, orderParams.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Error("failed to check if guest exists", zap.Error(err))
		return uuid.UUID{}, fmt.Errorf("failed to check if guest exists: %w", err)
	}

	if existingGuest != nil {
		return existingGuest.ID, nil
	}

	guestParams := &database.CreateGuestCheckoutParams{
		Email:         orderParams.Email,
		FirstName:     orderParams.FirstName,
		LastName:      orderParams.LastName,
		Phone:         sql.NullString{String: orderParams.Phone, Valid: true},
		StreetAddress: orderParams.StreetAddress,
		City:          orderParams.City,
		State:         orderParams.State,
		Country:       orderParams.Country,
	}
	newGuestID, err := s.guestCheckoutRepo.CreateGuestCheckout(ctx, guestParams)
	if err != nil {
		s.logger.Error("failed to create guest checkout", zap.Error(err))
		return uuid.UUID{}, fmt.Errorf("failed to create guest checkout: %w", err)
	}
	return *newGuestID, nil
}

func (s *OrderService) createOrderRecord(ctx context.Context, orderParams *model.CreateOrderParams, userUUID uuid.NullUUID) (uuid.UUID, error) {
	orderParam := &database.CreateOrderParams{
		UserID:          userUUID,
		GuestCheckoutID: orderParams.GuestCheckoutID,
		Total:           strconv.FormatFloat(orderParams.Total, 'f', -1, 64),
	}
	orderID, err := s.orderRepository.CreateOrder(ctx, orderParam)
	if err != nil {
		s.logger.Error("failed to create order", zap.Error(err))
		return uuid.UUID{}, fmt.Errorf("failed to create order: %w", err)
	}
	return *orderID, nil
}

func (s *OrderService) createOrderItem(ctx context.Context, orderID uuid.UUID, item model.CreateOrderItemParams) error {
	orderItem := &database.CreateOrderItemParams{
		OrderID:   uuid.NullUUID{UUID: orderID, Valid: true},
		ProductID: uuid.NullUUID{UUID: item.ProductID, Valid: true},
		Quantity:  item.Quantity,
		Price:     item.Price,
	}
	orderItemID, err := s.orderItemRepository.CreateOrderItem(ctx, orderItem)
	if err != nil {
		s.logger.Error("failed to create order item", zap.Error(err))
		return fmt.Errorf("failed to create order item: %w", err)
	}

	if err := s.createOrderItemOptions(ctx, orderItemID, item); err != nil {
		return err
	}

	return nil
}

func (s *OrderService) createOrderItemOptions(ctx context.Context, orderItemID *uuid.UUID, item model.CreateOrderItemParams) error {
	for _, optionID := range item.ProductOptionIDs {
		if optionID.Valid {
			if err := s.createOrderItemOption(ctx, *orderItemID, optionID.UUID); err != nil {
				return err
			}
		}
	}

	if item.ColorID.Valid {
		if err := s.createOrderItemOption(ctx, *orderItemID, item.ColorID.UUID); err != nil {
			return err
		}
	}

	if item.SizeID.Valid {
		if err := s.createOrderItemOption(ctx, *orderItemID, item.SizeID.UUID); err != nil {
			return err
		}
	}

	return nil
}

func (s *OrderService) createOrderItemOption(ctx context.Context, orderItemID, optionID uuid.UUID) error {
	optionValue, err := s.orderItemRepository.GetProductOptionValueByID(ctx, optionID)
	if err != nil {
		s.logger.Error("failed to get product option", zap.Error(err))
		return fmt.Errorf("failed to get product option: %w", err)
	}

	option, err := s.orderItemRepository.GetProductOptionByID(ctx, optionValue.OptionID.UUID)
	if err != nil {
		s.logger.Error("failed to get product option", zap.Error(err))
		return fmt.Errorf("failed to get product option: %w", err)
	}

	orderItemOption := &database.CreateOrderItemOptionParams{
		OrderItemID:     uuid.NullUUID{UUID: orderItemID, Valid: true},
		OptionType:      option.OptionName,
		OptionValue:     optionValue.ValueName,
		AdditionalPrice: optionValue.AdditionalPrice,
	}

	if err := s.orderItemRepository.CreateOrderItemOption(ctx, orderItemOption); err != nil {
		s.logger.Error("failed to create order item option", zap.Error(err))
		return fmt.Errorf("failed to create order item option: %w", err)
	}

	return nil
}

func (s *OrderService) createPayment(ctx context.Context, orderID uuid.UUID, total float64, paymentMethod string) (*model.OrderPayment, error) {
	paymentMethodID := getPaymentMethodID(paymentMethod)

	payment := &database.CreatePaymentParams{
		OrderID:           orderID,
		Amount:            strconv.FormatFloat(total, 'f', -1, 64),
		PaymentMethodID:   sql.NullInt32{Int32: paymentMethodID, Valid: true},
		PaymentStatusID:   sql.NullInt32{Int32: 1, Valid: true},
		CheckoutRequestID: "",
	}

	orderPayment, err := s.paymentRepository.CreatePayment(ctx, payment)
	if err != nil {
		s.logger.Error("failed to create payment", zap.Error(err), zap.Any("paymentParams", payment))
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return orderPayment, nil
}

func getPaymentMethodID(paymentMethod string) int32 {
	switch paymentMethod {
	case "delivery":
		return 1
	case "now":
		return 2
	default:
		return 1
	}
}

// ListOrders lists all orders
func (s *OrderService) ListOrders(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	// Check if user exists
	// user, err := s.userRepo.GetUserByID(ctx, userID)
	// if err != nil {
	// 	s.logger.Fatal("failed to get user", zap.Error(err))
	// 	return nil, fmt.Errorf("failed to get user: %w", err)
	// }

	//TODO: Check if user is admin

	orderIDs, err := s.orderRepository.GetOrderIDsByUserID(ctx, userID)
	if err != nil {
		s.logger.Fatal("failed to list orders", zap.Error(err))
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	return orderIDs, nil

}

// GetOrder gets order by ID
func (s *OrderService) GetOrder(ctx context.Context, orderID uuid.UUID) (*model.OrderClientResponse, error) {
	// Channels for results
	orderChan := make(chan *database.GetOrderByIdRow, 1)
	credentialsChan := make(chan *database.GetUserOrGuestCheckoutNameByOrderIDRow, 1)
	errChan := make(chan error, 2)

	// Ensure all channels are closed properly
	defer close(orderChan)
	defer close(credentialsChan)
	defer close(errChan)

	// Fetch order details in a separate goroutine
	go func() {
		order, err := s.orderRepository.GetOrderById(ctx, orderID)
		if err != nil {
			errChan <- fmt.Errorf("failed to get order: %w", err)
			return
		}
		orderChan <- order
	}()

	// Fetch customer name in a separate goroutine
	go func() {
		credentials, err := s.orderRepository.GetUserOrGuestCheckoutNameByOrderID(ctx, orderID)
		if err != nil {
			errChan <- fmt.Errorf("failed to get customer name: %w", err)
			return
		}
		credentialsChan <- credentials
	}()

	var order *database.GetOrderByIdRow
	var customerCredentials *database.GetUserOrGuestCheckoutNameByOrderIDRow

	// Wait for both operations to complete or an error to occur
	for i := 0; i < 2; i++ {
		select {
		case o := <-orderChan:
			order = o
		case n := <-credentialsChan:
			customerCredentials = n
		case err := <-errChan:
			s.logger.Error("operation failed", zap.Error(err))
			return nil, err
		case <-ctx.Done():
			s.logger.Error("operation cancelled", zap.Error(ctx.Err()))
			return nil, ctx.Err()
		}
	}

	// Check if user or guest
	var customerName string
	var phone string
	if customerCredentials.UserFirstName.Valid {
		customerName = customerCredentials.UserFirstName.String + " " + customerCredentials.UserLastName.String
		phone = customerCredentials.UserPhoneNumber.String
	} else {
		customerName = customerCredentials.GuestFirstName.String + " " + customerCredentials.GuestLastName.String
		phone = customerCredentials.GuestPhone.String
	}

	amount, err := strconv.ParseFloat(order.Total, 64)
	if err != nil {
		s.logger.Error("failed to parse amount", zap.Error(err))
		return nil, fmt.Errorf("failed to parse amount: %w", err)
	}

	// Create the response
	response := &model.OrderClientResponse{
		ID:             order.ID,
		OrderNumber:    order.OrderNumber.String,
		OrderCreatedAt: order.CreatedAt.Time,
		CustomerName:   customerName,
		Phone:          phone,
		Amount:         amount,
	}

	s.logger.Info("Order details", zap.Any("order", response))

	return response, nil
}

// CancelOrder cancels an order
func (s *OrderService) CancelOrder(ctx context.Context, orderID uuid.UUID, reason string) error {
	orderNumber, err := s.orderRepository.CancelOrder(ctx, orderID)
	if err != nil {
		s.logger.Error("failed to cancel order", zap.Error(err))
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	// Send order cancellation notification
	if err := utils.SendOrderCancellationNotification(s.cfg, orderNumber, reason); err != nil {
		s.logger.Error("failed to send order cancellation notification", zap.Error(err))
		return fmt.Errorf("failed to send order cancellation notification: %w", err)
	}

	return nil
}

// ChangeOrderPaymentMethod changes the payment method of an order
func (s *OrderService) ChangeOrderPaymentMethod(ctx context.Context, orderID uuid.UUID, method string) error {
	orderNumber, err := s.paymentRepository.ChangeOrderPaymentMethod(ctx, orderID, method)
	if err != nil {
		s.logger.Error("failed to change payment method", zap.Error(err))
		return fmt.Errorf("failed to change payment method: %w", err)
	}

	// Send order payment method change notification
	if err := utils.SendOrderPaymentMethodChangeNotification(s.cfg, orderNumber, method); err != nil {
		s.logger.Error("failed to send order payment method change notification", zap.Error(err))
		return fmt.Errorf("failed to send order payment method change notification: %w", err)
	}

	return nil
}
