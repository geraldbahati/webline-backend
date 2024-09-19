package services

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"
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
	discountRepo        *repository.DiscountRepository	
	exchangeRateRepo    repository.ExchangeRateRepository
	companyRepo         repository.CompanyRepository
	cfg                 *appconfig.Config
}

func NewOrderService(logger *zap.Logger, guestCheckoutRepo *repository.GuestCheckoutRepository, orderRepository *repository.OrderRepository, orderItemRepository *repository.OrderItemRepository, paymentRepository *repository.PaymentRepository, userRepo *repository.UserRepository, productRepo *repository.ProductRepository, discountRepo *repository.DiscountRepository, exchangeRateRepo repository.ExchangeRateRepository, companyRepo repository.CompanyRepository, cfg *appconfig.Config) *OrderService {
	return &OrderService{
		logger:              logger,
		guestCheckoutRepo:   guestCheckoutRepo,
		orderRepository:     orderRepository,
		orderItemRepository: orderItemRepository,
		paymentRepository:   paymentRepository,
		userRepo:            userRepo,
		productRepo:         productRepo,
		exchangeRateRepo:    exchangeRateRepo,
		companyRepo:         companyRepo,
		discountRepo:        discountRepo,
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

// Optimized CreateOrder creates a new order with items
func (s *OrderService) CreateOrder(ctx context.Context, orderParams *model.CreateOrderParams, items []model.CreateOrderItemParams) (*uuid.UUID, error) {
	var orderAmounts model.OrderAmounts
	var orderID uuid.UUID
	var notificationItems []utils.OrderItem

	err := s.orderRepository.ExecTx(ctx, func(q *database.Queries) error {
		// Create the order record
		orderID, err := s.createOrderRecord(ctx, q, orderParams)
		if err != nil {
			return fmt.Errorf("createOrderRecord failed: %w", err)
		}

		// Process the order items
		notifyItems, totalPrice, totalDiscount, err := s.processOrderItems(ctx, q, orderID, items)
		if err != nil {
			return fmt.Errorf("processOrderItems failed: %w", err)
		}

		notificationItems = notifyItems

		// Update order amounts in the database
		updateParams := database.UpdateOrderAmountsParams{
			Column1:        orderID,
			Column2:     fmt.Sprintf("%.2f", totalPrice),
			Column3:        "0", 
			Column4:        "0", 
			Column5:  fmt.Sprintf("%.2f", totalDiscount),
		}

		row, err := q.UpdateOrderAmounts(ctx, updateParams)
		if err != nil {
			return fmt.Errorf("UpdateOrderAmounts failed: %w", err)
		}

		// Parse the float fields directly without using intermediate maps
		orderAmounts, err = parseOrderAmounts(row, s.logger)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		s.logger.Error("failed to create order", zap.Error(err))
		return nil, err
	}

	// Update the order parameters with the calculated amounts
	updateOrderParams(orderParams, &orderAmounts)

	// Send order notification asynchronously
	go s.sendOrderNotification(ctx, orderID, orderParams, notificationItems)

	return &orderID, nil
}

// parseOrderAmounts parses the order amount fields from the database row
func parseOrderAmounts(row database.UpdateOrderAmountsRow, logger *zap.Logger) (model.OrderAmounts, error) {
	var orderAmounts model.OrderAmounts
	var err error

	orderAmounts.SubTotal, err = parseFloatField("subtotal", row.Subtotal, logger)
	if err != nil {
		return orderAmounts, err
	}

	orderAmounts.TaxAmount, err = parseFloatField("tax amount", row.TaxAmount, logger)
	if err != nil {
		return orderAmounts, err
	}

	orderAmounts.ShippingAmount, err = parseFloatField("shipping amount", row.ShippingAmount, logger)
	if err != nil {
		return orderAmounts, err
	}

	orderAmounts.DiscountAmount, err = parseFloatField("discount amount", row.DiscountAmount, logger)
	if err != nil {
		return orderAmounts, err
	}

	orderAmounts.VatAmount, err = parseFloatField("vat amount", row.VatAmount, logger)
	if err != nil {
		return orderAmounts, err
	}

	orderAmounts.GrandTotal, err = parseFloatField("grand total", row.GrandTotal, logger)
	if err != nil {
		return orderAmounts, err
	}

	return orderAmounts, nil
}

// parseFloatField parses a string field to float64 and logs any errors
func parseFloatField(fieldName, fieldValue string, logger *zap.Logger) (float64, error) {
	value, err := strconv.ParseFloat(fieldValue, 64)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to parse %s", fieldName), zap.Error(err))
		return 0, fmt.Errorf("parse %s: %w", fieldName, err)
	}
	return value, nil
}

// updateOrderParams updates the orderParams with the calculated orderAmounts
func updateOrderParams(orderParams *model.CreateOrderParams, orderAmounts *model.OrderAmounts) {
	orderParams.GrandTotal = orderAmounts.GrandTotal
	orderParams.SubTotal = orderAmounts.SubTotal
	orderParams.DiscountAmount = orderAmounts.DiscountAmount
	orderParams.VatAmount = orderAmounts.VatAmount
}

func (s *OrderService) sendOrderNotification(ctx context.Context, orderID uuid.UUID, orderParams *model.CreateOrderParams, items []utils.OrderItem) {
	// Send the order notification using the utils package
	err := utils.SendOrderNotification(s.cfg, orderID, orderParams, items)
	if err != nil {
		s.logger.Error("failed to send order notification", zap.Error(err))
	}
}

func (s *OrderService) processOrderItems(ctx context.Context, q *database.Queries, orderID uuid.UUID, items []model.CreateOrderItemParams) ([]utils.OrderItem, float64, float64, error) {
    var totalPriceKES float64
    var totalDiscountAmount float64
	var notificationItems []utils.OrderItem

    // Fetch exchange rate once
    exchangeRate, err := s.exchangeRateRepo.GetLatestExchangeRate(ctx, "USD")
    if err != nil {
        return nil, 0, 0, fmt.Errorf("failed to get exchange rate: %w", err)
    }

    // Collect product IDs from items
    productIDs := make([]uuid.UUID, len(items))
    for i, item := range items {
        productIDs[i] = item.ProductID
    }

    // Fetch all products at once
    products, err := s.productRepo.GetProductsByIDs(ctx, productIDs)
    if err != nil {
        return nil, 0, 0, fmt.Errorf("failed to get products: %w", err)
    }

    // Build a map of products for quick lookup
    productMap := make(map[uuid.UUID]model.ProductSchema)
    for _, product := range products {
        productMap[product.ID] = product
    }

    // Fetch applicable discounts for products
    discounts, err := s.discountRepo.GetActiveDiscountsByProductIDs(ctx, productIDs, time.Now())
    if err != nil {
        return nil, 0, 0, fmt.Errorf("failed to get discounts: %w", err)
    }

    // Build a map of discounts
    discountMap := make(map[uuid.UUID]float64) // productID -> discountPercentage
    for _, discount := range discounts {
        discountMap[discount.ProductID] = float64(discount.DiscountPercentage)
    }

    // Prepare data for batch insertion
    orderIDs := make([]uuid.UUID, len(items))
    productIDsArray := make([]uuid.UUID, len(items))
    quantities := make([]int32, len(items))
    productNames := make([]string, len(items))
    productSkus := make([]string, len(items))
    unitPrices := make([]string, len(items))
    discountAmounts := make([]string, len(items))
    totalPrices := make([]string, len(items))

    for i, item := range items {
        product, ok := productMap[item.ProductID]
        if !ok {
            return nil, 0, 0, fmt.Errorf("product not found: %s", item.ProductID)
        }

        // Get the price in USD and convert to KES
        priceUSD, err := strconv.ParseFloat(product.USD, 64)
        if err != nil {
            return nil, 0, 0, fmt.Errorf("failed to parse product price: %w", err)
        }
        priceKES := priceUSD * exchangeRate

        // Calculate the unit price considering any additional prices from options
        unitPrice := priceKES

        // Here you might need to add the price of any selected options
        // For simplicity, we'll assume unitPrice is priceKES

        // Check for discount
        discountPercentage, hasDiscount := discountMap[item.ProductID]
        var discountAmount float64
        if hasDiscount {
            discountAmount = unitPrice * float64(item.Quantity) * (discountPercentage / 100)
        } else {
            discountAmount = 0
        }

        itemTotal := (unitPrice * float64(item.Quantity)) - discountAmount
        if itemTotal < 0 {
            itemTotal = 0 // Ensure total doesn't go negative
        }
        totalPriceKES += itemTotal
        totalDiscountAmount += discountAmount

        orderIDs[i] = orderID
        productIDsArray[i] = item.ProductID
        quantities[i] = item.Quantity
        productNames[i] = product.Name
        productSkus[i] = product.Slug
        unitPrices[i] = fmt.Sprintf("%.2f", unitPrice)
        discountAmounts[i] = fmt.Sprintf("%.2f", discountAmount)
        totalPrices[i] = fmt.Sprintf("%.2f", itemTotal)

		notificationItems = append(notificationItems, utils.OrderItem{
			ProductName: product.Name,
			Quantity:    item.Quantity,
			Price:       itemTotal,
		})
    }

    orderItemsParam := database.CreateOrderItemsParams{
        Column1:        orderIDs,
        Column2:      productIDsArray,
        Column3:      quantities,
        Column4:      productNames,
        Column5:      productSkus,
        Column6:      unitPrices,
        Column7:      discountAmounts,
        Column8:      totalPrices,
    }

    // Batch insert order items
    if err := q.CreateOrderItems(ctx, orderItemsParam); err != nil {
        return nil, 0, 0, fmt.Errorf("failed to create order items: %w", err)
    }

    // Apply rounding to totalPriceKES
    roundedTotalPriceKES := utils.RoundPrice(totalPriceKES)

    return notificationItems, roundedTotalPriceKES, totalDiscountAmount, nil
}




func (s *OrderService) createOrderRecord(ctx context.Context, q *database.Queries, orderParams *model.CreateOrderParams) (uuid.UUID, error) {
	// create company if it company name and kra pin are provided
	var companyID *uuid.UUID
	var err error
	if *orderParams.CompanyName != "" && *orderParams.KraPIN != "" {
		companyID, err = s.companyRepo.CreateCompany(ctx, *orderParams.CompanyName, *orderParams.KraPIN, orderParams.County, orderParams.Phone, orderParams.Email)
		if err != nil && err != sql.ErrNoRows {
			s.logger.Error("failed to create company", zap.Error(err))
			return uuid.UUID{}, fmt.Errorf("failed to create company: %w", err)
		}

		if err == sql.ErrNoRows {
			// get company id
			companyID, err = s.companyRepo.GetCompanyID(ctx, *orderParams.CompanyName, *orderParams.KraPIN)
			if err != nil {
				s.logger.Error("failed to get company ID", zap.Error(err))
				return uuid.UUID{}, fmt.Errorf("failed to get company ID: %w", err)
			}
		}
	}

	orderParam := database.CreateOrderParams{
		GuestCheckoutID: uuid.NullUUID{UUID: *orderParams.GuestID, Valid: orderParams.GuestID != nil},
		CompanyName:     sql.NullString{String: *orderParams.CompanyName, Valid: orderParams.CompanyName != nil},
		UserID:          uuid.NullUUID{UUID: *orderParams.UserID, Valid: orderParams.UserID != nil},
		CompanyID:       uuid.NullUUID{UUID: *companyID, Valid: companyID != nil},
		Column6:         "KES",
		KraPin:          sql.NullString{String: *orderParams.KraPIN, Valid: orderParams.KraPIN != nil},
	}
	orderID, err := q.CreateOrder(ctx, orderParam)
	if err != nil {
		s.logger.Error("failed to create order", zap.Error(err))
		return uuid.UUID{}, fmt.Errorf("failed to create order: %w", err)
	}
	return orderID, nil
}

func (s *OrderService) createOrderItem(ctx context.Context, orderID uuid.UUID, item model.CreateOrderItemParams) (float64, error) {
	product, err := s.productRepo.GetProductByID(ctx, item.ProductID)
	if err != nil {
		s.logger.Error("failed to get product", zap.Error(err))
		return 0, fmt.Errorf("failed to get product: %w", err)
	}

	convertedPrice, err := s.convertPriceToKES(ctx, product.USD)
	if err != nil {
		s.logger.Error("failed to convert price", zap.Error(err))
		return 0, fmt.Errorf("failed to convert price: %w", err)
	}

	totalPrice := convertedPrice * float64(item.Quantity)

	orderItem := &database.CreateOrderItemParams{
		OrderID:     uuid.NullUUID{UUID: orderID, Valid: true},
		ProductID:   uuid.NullUUID{UUID: item.ProductID, Valid: true},
		Quantity:    item.Quantity,
		ProductName: product.Name,
		ProductSku:  product.Slug,
		UnitPrice:   fmt.Sprintf("%.2f", convertedPrice),
		TotalPrice:  fmt.Sprintf("%.2f", totalPrice),
	}
	_, err = s.orderItemRepository.CreateOrderItem(ctx, orderItem)
	if err != nil {
		s.logger.Error("failed to create order item", zap.Error(err))
		return 0, fmt.Errorf("failed to create order item: %w", err)
	}


	return totalPrice, nil
}

func (s *OrderService) prepareOrderItems(ctx context.Context, items []model.CreateOrderItemParams) ([]utils.OrderItem, error) {
	orderItems := make([]utils.OrderItem, 0, len(items))

	for _, item := range items {
		product, err := s.productRepo.GetProductByID(ctx, item.ProductID)
		if err != nil {
			s.logger.Error("failed to get product", zap.Error(err))
			return nil, fmt.Errorf("failed to get product: %w", err)
		}

		convertedPrice, err := s.convertPriceToKES(ctx, product.USD)
		if err != nil {
			s.logger.Error("failed to convert price", zap.Error(err))
			return nil, fmt.Errorf("failed to convert price: %w", err)
		}

		orderItems = append(orderItems, utils.OrderItem{
			ProductName: product.Name,
			Quantity:    item.Quantity,
			Price:       convertedPrice,
		})
	}
	return orderItems, nil
}

func (s *OrderService) convertPriceToKES(ctx context.Context, price string) (float64, error) {
	exchangeRate, err := s.exchangeRateRepo.GetLatestExchangeRate(ctx, "USD")
	if err != nil {
		s.logger.Error("failed to get exchange rate", zap.Error(err))
		return 0, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	priceFloat, err := strconv.ParseFloat(price, 64)
	if err != nil {
		s.logger.Error("failed to parse price", zap.Error(err))
		return 0, fmt.Errorf("failed to parse price: %w", err)
	}

	convertedPrice := priceFloat * exchangeRate
	return convertedPrice, nil
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

	amount, err := strconv.ParseFloat(order.GrandTotal, 64)
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

// GetTotalRevenue gets the total revenue
func (s *OrderService) GetTotalRevenue(ctx context.Context) (*model.Revenue, error) {
	statusID, err := s.paymentRepository.GetPaymentStatusIDByStatus(ctx, "paid")
	if err != nil {
		s.logger.Error("failed to get payment status ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get payment status ID: %w", err)
	}

	revenue, err := s.orderRepository.GetTotalRevenue(ctx, statusID)
	if err != nil {
		s.logger.Error("failed to get total revenue", zap.Error(err))
		return nil, fmt.Errorf("failed to get total revenue: %w", err)
	}

	currentRevenue, previousRevenue, err := s.orderRepository.GetTotalRevenueForLastTwoMonths(ctx, statusID)
	if err != nil {
		s.logger.Error("failed to get last two months revenue", zap.Error(err))
		return nil, fmt.Errorf("failed to get last two months revenue: %w", err)
	}

	// percentage growth
	percentageGrowth := 0.0
	if previousRevenue != 0 {
		percentageGrowth = ((currentRevenue - previousRevenue) / previousRevenue) * 100
	}

	result := &model.Revenue{
		Revenue:       revenue,
		MonthlyGrowth: percentageGrowth,
	}

	s.logger.Info("Total revenue", zap.Any("revenue", result))

	return result, nil
}

// GetMonthlySales gets the total revenue for the current month
func (s *OrderService) GetMonthlySales(ctx context.Context) (*model.Revenue, error) {
	statusID, err := s.paymentRepository.GetPaymentStatusIDByStatus(ctx, "paid")
	if err != nil {
		s.logger.Error("failed to get payment status ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get payment status ID: %w", err)
	}

	currentRevenue, lastRevenue, err := s.orderRepository.GetMonthlySalesForLastTwoMonths(ctx, statusID)
	if err != nil {
		s.logger.Error("failed to get last month revenue", zap.Error(err))
		return nil, fmt.Errorf("failed to get last month revenue: %w", err)
	}

	// percentage growth
	percentageGrowth := 0.0
	if lastRevenue != 0 {
		percentageGrowth = ((currentRevenue - lastRevenue) / lastRevenue) * 100
	}

	result := &model.Revenue{
		Revenue:       currentRevenue,
		MonthlyGrowth: percentageGrowth,
	}

	s.logger.Info("Monthly sales", zap.Any("revenue", result))
	return result, nil
}

// GetMonthlyRevenue gets the monthly revenue
func (s *OrderService) GetMonthlyRevenue(ctx context.Context) ([]*model.MonthlyRevenue, error) {
	statusID, err := s.paymentRepository.GetPaymentStatusIDByStatus(ctx, "paid")
	if err != nil {
		s.logger.Error("failed to get payment status ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get payment status ID: %w", err)
	}

	revenue, err := s.orderRepository.GetMonthlyRevenue(ctx, statusID)
	if err != nil {
		s.logger.Error("failed to get monthly revenue", zap.Error(err))
		return nil, fmt.Errorf("failed to get monthly revenue: %w", err)
	}

	s.logger.Info("Monthly revenue", zap.Any("revenue", revenue))
	return revenue, nil
}

// GetSalesTrend gets the sales trend
func (s *OrderService) GetSalesTrend(ctx context.Context) (float64, error) {
	statusID, err := s.paymentRepository.GetPaymentStatusIDByStatus(ctx, "paid")
	if err != nil {
		s.logger.Error("failed to get payment status ID", zap.Error(err))
		return 0, fmt.Errorf("failed to get payment status ID: %w", err)
	}

	currentSales, lastSales, err := s.orderRepository.GetSalesTrend(ctx, statusID)
	if err != nil {
		s.logger.Error("failed to get sales trend", zap.Error(err))
		return 0, fmt.Errorf("failed to get sales trend: %w", err)
	}

	// percentage growth
	percentageGrowth := 0.0
	if lastSales != 0 {
		percentageGrowth = ((currentSales - lastSales) / lastSales) * 100
	}

	s.logger.Info("Sales trend", zap.Float64("sales", percentageGrowth))
	return percentageGrowth, nil
}

// GetRecentSales gets the recent sales
func (s *OrderService) GetRecentSales(ctx context.Context) ([]*model.OrderUser, error) {

	recentSales, err := s.orderRepository.GetRecentSales(ctx)
	if err != nil {
		s.logger.Error("failed to get recent sales", zap.Error(err))
		return nil, fmt.Errorf("failed to get recent sales: %w", err)
	}

	s.logger.Info("Recent sales", zap.Any("sales", recentSales))
	return recentSales, nil
}

// GetTotalSalesCurrentMonth gets the total sales for the current month
func (s *OrderService) GetTotalSalesCurrentMonth(ctx context.Context) (int64, error) {
	sales, err := s.orderRepository.GetTotalSalesCurrentMonth(ctx)
	if err != nil {
		s.logger.Error("failed to get total sales for current month", zap.Error(err))
		return 0, fmt.Errorf("failed to get total sales for current month: %w", err)
	}

	s.logger.Info("Total sales for current month", zap.Int64("sales", sales))
	return sales, nil
}

// GetExchangeRate gets the exchange rate
func (s *OrderService) GetExchangeRate(ctx context.Context) (float64, error) {
	exchangeRate, err := s.exchangeRateRepo.GetLatestExchangeRate(ctx, "USD")
	if err != nil {
		s.logger.Error("failed to get exchange rate", zap.Error(err))
		return 0, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	s.logger.Info("Exchange rate", zap.Any("exchangeRate", exchangeRate))
	return exchangeRate, nil
}

// UpdateExchangeRate updates the exchange rate
func (s *OrderService) UpdateExchangeRate(ctx context.Context, rate float64) error {
	// Get today's date
	validFrom := time.Now()

	// valid date range is 30 days
	validTo := validFrom.AddDate(0, 0, 30)

	err := s.exchangeRateRepo.UpdateExchangeRate(ctx, "USD", rate, validFrom, validTo)
	if err != nil {
		s.logger.Error("failed to update exchange rate", zap.Error(err))
		return fmt.Errorf("failed to update exchange rate: %w", err)
	}

	return nil
}
