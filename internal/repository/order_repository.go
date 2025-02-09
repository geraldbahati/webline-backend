package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrderRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewOrderRepository(db *sql.DB, logger *zap.Logger) *OrderRepository {
	return &OrderRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// ExecTx executes a database transaction with the provided function
func (r *OrderRepository) ExecTx(ctx context.Context, fn func(*database.Queries) error) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	q := database.New(tx)
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			r.logger.Panic("transaction panicked, rolling back", zap.Any("panic", p))
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				r.logger.Error("rollback failed", zap.Error(rbErr))
				err = fmt.Errorf("rollback transaction: %w", rbErr)
			} else {
				r.logger.Warn("transaction rolled back due to error", zap.Error(err))
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				r.logger.Error("commit failed", zap.Error(commitErr))
				err = fmt.Errorf("commit transaction: %w", commitErr)
			}
		}
	}()
	err = fn(q)
	return err
}

// CreateOrder creates a new order
func (r *OrderRepository) CreateOrder(ctx context.Context, order *database.CreateOrderParams) (*uuid.UUID, error) {
	var orderID uuid.UUID
	if err := r.ExecTx(ctx, func(q *database.Queries) error {
		orderRecord, err := q.CreateOrder(ctx, *order)
		if err != nil {
			return fmt.Errorf("createOrderRecord failed: %w", err)
		}
		orderID = orderRecord.ID
		r.logger.Info("Order record created", zap.String("orderID", orderRecord.ID.String()))

		// Process order items and update order amounts here...
		// (Assume processOrderItems and updateOrderAmounts logic remains unchanged)
		return nil
	}); err != nil {
		r.logger.Error("Create order transaction failed", zap.Error(err))
		return nil, fmt.Errorf("create order transaction: %w", err)
	}
	r.logger.Info("Order created successfully", zap.String("orderID", orderID.String()))
	return &orderID, nil
}

// UpdateOrderStatus updates the status of an order
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error {
	if err := r.ExecTx(ctx, func(q *database.Queries) error {
		err := q.UpdateOrderStatus(ctx, database.UpdateOrderStatusParams{
			ID:     orderID,
			Status: status,
		})
		if err != nil {
			r.logger.Error("update order status failed", zap.Error(err))
			return fmt.Errorf("update order status: %w", err)
		}
		return nil
	}); err != nil {
		r.logger.Error("update order status transaction failed", zap.Error(err))
		return fmt.Errorf("update order status transaction: %w", err)
	}
	return nil
}

// UpdateOrderPaymentStatus updates the payment status of an order
func (r *OrderRepository) UpdateOrderPaymentStatus(ctx context.Context, orderID uuid.UUID, paymentStatus string) error {
	if err := r.ExecTx(ctx, func(q *database.Queries) error {
		err := q.UpdateOrderPaymentStatus(ctx, database.UpdateOrderPaymentStatusParams{
			ID:            orderID,
			PaymentStatus: paymentStatus,
		})
		if err != nil {
			r.logger.Error("update order payment status failed", zap.Error(err))
			return fmt.Errorf("update order payment status: %w", err)
		}
		return nil
	}); err != nil {
		r.logger.Error("update order payment status transaction failed", zap.Error(err))
		return fmt.Errorf("update order payment status transaction: %w", err)
	}
	return nil
}

// GetOrderById retrieves an order by its ID
func (r *OrderRepository) GetOrderById(ctx context.Context, orderID uuid.UUID) (*database.GetOrderByIdRow, error) {
	order, err := r.Queries.GetOrderById(ctx, orderID)
	if err != nil {
		r.logger.Error("get order by id failed", zap.Error(err))
		return nil, fmt.Errorf("get order by id: %w", err)
	}
	return &order, nil
}

// GetOrdersByUserId retrieves all orders for a specific user
func (r *OrderRepository) GetOrdersByUserId(ctx context.Context, userID uuid.NullUUID) ([]database.GetOrdersByUserIdRow, error) {
	orders, err := r.Queries.GetOrdersByUserId(ctx, userID)
	if err != nil {
		r.logger.Error("get orders by user id failed", zap.Error(err))
		return nil, fmt.Errorf("get orders by user id: %w", err)
	}
	return orders, nil
}

// GetOrdersByGuestCheckoutId retrieves all orders for a specific guest checkout
func (r *OrderRepository) GetOrdersByGuestCheckoutId(ctx context.Context, guestCheckoutID uuid.NullUUID) ([]database.GetOrdersByGuestCheckoutIdRow, error) {
	orders, err := r.Queries.GetOrdersByGuestCheckoutId(ctx, guestCheckoutID)
	if err != nil {
		r.logger.Error("get orders by guest checkout id failed", zap.Error(err))
		return nil, fmt.Errorf("get orders by guest checkout id: %w", err)
	}
	return orders, nil
}

// CreateGuestCheckout creates a guest checkout
func (r *OrderRepository) CreateGuestCheckout(ctx context.Context, guestCheckout *database.CreateGuestCheckoutParams) (*uuid.UUID, error) {
	var orderID uuid.UUID
	if err := r.ExecTx(ctx, func(q *database.Queries) error {
		var err error
		orderID, err = q.CreateGuestCheckout(ctx, *guestCheckout)
		if err != nil {
			r.logger.Error("create guest checkout failed", zap.Error(err))
			return fmt.Errorf("create guest checkout: %w", err)
		}
		return nil
	}); err != nil {
		r.logger.Error("create guest checkout transaction failed", zap.Error(err))
		return nil, fmt.Errorf("create guest checkout transaction: %w", err)
	}
	return &orderID, nil
}

// GetOrderIDsByUserID retrieves all order IDs for a specific user
func (r *OrderRepository) GetOrderIDsByUserID(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	orderIDs, err := r.Queries.GetOrderIDsByUserID(ctx, uuid.NullUUID{UUID: userID, Valid: true})
	if err != nil {
		r.logger.Error("get order IDs by user ID failed", zap.Error(err))
		return nil, fmt.Errorf("get order IDs by user ID: %w", err)
	}

	r.logger.Info("Order IDs retrieved successfully", zap.Any("orderIDs", orderIDs))
	return orderIDs, nil
}

// GetUserOrGuestCheckoutNameByOrderID retrieves the name of the user or guest checkout for a specific order
func (r *OrderRepository) GetUserOrGuestCheckoutNameByOrderID(ctx context.Context, orderID uuid.UUID) (*database.GetUserOrGuestCheckoutNameByOrderIDRow, error) {
	names, err := r.Queries.GetUserOrGuestCheckoutNameByOrderID(ctx, orderID)
	if err != nil {
		r.logger.Error("get user or guest checkout name by order ID failed", zap.Error(err))
		return nil, fmt.Errorf("get user or guest checkout name by order ID: %w", err)
	}

	r.logger.Info("User or guest checkout name retrieved successfully", zap.Any("names", names))
	return &names, nil
}

// CancelOrder cancels an order
func (r *OrderRepository) CancelOrder(ctx context.Context, orderID uuid.UUID) (string, error) {
	var orderNumber string
	if err := r.ExecTx(ctx, func(q *database.Queries) error {
		orderNum, err := q.CancelOrder(ctx, orderID)
		if err != nil {
			r.logger.Error("cancel order failed", zap.Error(err))
			return fmt.Errorf("cancel order: %w", err)
		}
		orderNumber = orderNum.String
		return nil
	}); err != nil {
		r.logger.Error("cancel order transaction failed", zap.Error(err))
		return "", fmt.Errorf("cancel order transaction: %w", err)
	}

	r.logger.Info("Order cancelled successfully", zap.String("orderNumber", orderNumber))
	return orderNumber, nil
}

// GetTotalRevenue retrieves the total revenue
func (r *OrderRepository) GetTotalRevenue(ctx context.Context, paymentStatus int32) (float64, error) {
	revenue, err := r.Queries.GetTotalRevenueByPaymentStatus(ctx, sql.NullInt32{
		Int32: paymentStatus,
		Valid: true,
	})
	if err != nil {
		r.logger.Error("get total revenue failed", zap.Error(err))
		return 0, fmt.Errorf("get total revenue: %w", err)
	}

	total, err := strconv.ParseFloat(revenue, 64)
	if err != nil {
		r.logger.Error("failed to parse total revenue", zap.Error(err))
		return 0, fmt.Errorf("parse total revenue: %w", err)
	}

	r.logger.Info("Total revenue retrieved successfully", zap.Float64("revenue", total))
	return total, nil
}

// GetMonthlySales retrieves the monthly sales
func (r *OrderRepository) GetMonthlySales(ctx context.Context, paymentStatus int32) ([]database.GetMonthlySalesRow, error) {
	sales, err := r.Queries.GetMonthlySales(ctx, sql.NullInt32{
		Int32: paymentStatus,
		Valid: true,
	})
	if err != nil {
		r.logger.Error("get monthly sales failed", zap.Error(err))
		return nil, fmt.Errorf("get monthly sales: %w", err)
	}

	r.logger.Info("Monthly sales retrieved successfully", zap.Any("sales", sales))
	return sales, nil
}

// GetTotalRevenueForLastTwoMonths retrieves the total revenue for the last two months
func (r *OrderRepository) GetTotalRevenueForLastTwoMonths(ctx context.Context, paymentStatus int32) (float64, float64, error) {
	revenue, err := r.Queries.GetTotalRevenueForLastTwoMonths(ctx, sql.NullInt32{
		Int32: paymentStatus,
		Valid: true,
	})
	if err != nil {
		r.logger.Error("get total revenue for last two months failed", zap.Error(err))
		return 0, 0, fmt.Errorf("get total revenue for last two months: %w", err)
	}

	currentRevenue, err := strconv.ParseFloat(revenue.CurrentMonthRevenue, 64)
	if err != nil {
		r.logger.Error("failed to parse total revenue for last two months", zap.Error(err))
		return 0, 0, fmt.Errorf("parse total revenue for last two months: %w", err)
	}

	lastRevenue, err := strconv.ParseFloat(revenue.PreviousMonthRevenue, 64)
	if err != nil {
		r.logger.Error("failed to parse total revenue for last two months", zap.Error(err))
		return 0, 0, fmt.Errorf("parse total revenue for last two months: %w", err)
	}

	r.logger.Info("Total revenue for last two months retrieved successfully", zap.Float64("currentRevenue", currentRevenue), zap.Float64("lastRevenue", lastRevenue))
	return currentRevenue, lastRevenue, nil
}

// GetMonthlySalesForLastTwoMonths retrieves the monthly sales for the last two months
func (r *OrderRepository) GetMonthlySalesForLastTwoMonths(ctx context.Context, paymentStatus int32) (float64, float64, error) {
	sales, err := r.Queries.GetMonthlySalesForLastTwoMonths(ctx, sql.NullInt32{
		Int32: paymentStatus,
		Valid: true,
	})
	if err != nil {
		r.logger.Error("get monthly sales for last two months failed", zap.Error(err))
		return 0, 0, fmt.Errorf("get monthly sales for last two months: %w", err)
	}

	currentSales, err := strconv.ParseFloat(sales.CurrentMonthSales, 64)
	if err != nil {
		r.logger.Error("failed to parse monthly sales for last two months", zap.Error(err))
		return 0, 0, fmt.Errorf("parse monthly sales for last two months: %w", err)
	}

	lastSales, err := strconv.ParseFloat(sales.PreviousMonthSales, 64)
	if err != nil {
		r.logger.Error("failed to parse monthly sales for last two months", zap.Error(err))
		return 0, 0, fmt.Errorf("parse monthly sales for last two months: %w", err)
	}

	r.logger.Info("Monthly sales for last two months retrieved successfully", zap.Float64("currentSales", currentSales), zap.Float64("lastSales", lastSales))
	return currentSales, lastSales, nil
}

// GetMonthlyRevenue retrieves the monthly revenue
func (r *OrderRepository) GetMonthlyRevenue(ctx context.Context, paymentStatus int32) ([]*model.MonthlyRevenue, error) {
	revenue, err := r.Queries.GetMonthlyRevenue(ctx, sql.NullInt32{
		Int32: paymentStatus,
		Valid: true,
	})
	if err != nil {
		r.logger.Error("get monthly revenue failed", zap.Error(err))
		return nil, fmt.Errorf("get monthly revenue: %w", err)
	}

	var monthlyRevenue []*model.MonthlyRevenue
	for _, row := range revenue {
		totalSales, err := strconv.ParseFloat(row.TotalSales, 64)
		if err != nil {
			r.logger.Error("failed to parse monthly revenue", zap.Error(err))
			return nil, fmt.Errorf("parse monthly revenue: %w", err)
		}

		monthlyRevenue = append(monthlyRevenue, &model.MonthlyRevenue{
			Month:   time.Unix(row.Month, 0),
			Revenue: totalSales,
		})
	}

	r.logger.Info("Monthly revenue retrieved successfully", zap.Any("monthlyRevenue", monthlyRevenue))
	return monthlyRevenue, nil
}

// GetSalesTrend retrieves the sales trend
func (r *OrderRepository) GetSalesTrend(ctx context.Context, paymentStatus int32) (float64, float64, error) {
	revenue, err := r.Queries.GetSalesTrend(ctx, sql.NullInt32{
		Int32: paymentStatus,
		Valid: true,
	})
	if err != nil {
		r.logger.Error("get sales trend failed", zap.Error(err))
		return 0, 0, fmt.Errorf("get sales trend: %w", err)
	}

	currentSales, err := strconv.ParseFloat(revenue.CurrentMonthSales, 64)
	if err != nil {
		r.logger.Error("failed to parse sales trend", zap.Error(err))
		return 0, 0, fmt.Errorf("parse sales trend: %w", err)
	}

	lastSales, err := strconv.ParseFloat(revenue.PreviousMonthSales, 64)
	if err != nil {
		r.logger.Error("failed to parse sales trend", zap.Error(err))
		return 0, 0, fmt.Errorf("parse sales trend: %w", err)
	}

	r.logger.Info("Sales trend retrieved successfully", zap.Float64("currentSales", currentSales), zap.Float64("lastSales", lastSales))
	return currentSales, lastSales, nil
}

// GetRecentSales retrieves the recent sales
func (r *OrderRepository) GetRecentSales(ctx context.Context) ([]*model.OrderUser, error) {
	sales, err := r.Queries.GetRecentSales(ctx)
	if err != nil {
		r.logger.Error("get recent sales failed", zap.Error(err))
		return nil, fmt.Errorf("get recent sales: %w", err)
	}

	var recentSales []*model.OrderUser
	for _, row := range sales {
		sale, err := strconv.ParseFloat(row.Amount, 64)
		if err != nil {
			r.logger.Error("failed to parse recent sales", zap.Error(err))
			return nil, fmt.Errorf("parse recent sales: %w", err)
		}

		recentSales = append(recentSales, &model.OrderUser{
			Name:     row.Name,
			Email:    row.Email,
			Amount:   sale,
			Fallback: row.Fallback,
		})
	}

	r.logger.Info("Recent sales retrieved successfully", zap.Any("recentSales", recentSales))
	return recentSales, nil
}

// GetTotalSalesCurrentMonth retrieves the total sales for the current month
func (r *OrderRepository) GetTotalSalesCurrentMonth(ctx context.Context) (int64, error) {
	revenue, err := r.Queries.GetTotalSalesCurrentMonth(ctx)
	if err != nil {
		r.logger.Error("get total sales for current month failed", zap.Error(err))
		return 0, fmt.Errorf("get total sales for current month: %w", err)
	}

	r.logger.Info("Total sales for current month retrieved successfully", zap.Int64("revenue", revenue))
	return revenue, nil
}

// Start of Selection
// UpdateOrderAmounts updates the amounts for an order
func (r *OrderRepository) UpdateOrderAmounts(ctx context.Context, orderID uuid.UUID, subtotal, taxAmount, shippingAmount, discountAmount float64) (*model.OrderAmounts, error) {
	var orderAmounts model.OrderAmounts

	err := r.ExecTx(ctx, func(q *database.Queries) error {
		row, err := q.UpdateOrderAmounts(ctx, database.UpdateOrderAmountsParams{
			Column1: orderID,
			Column2: fmt.Sprintf("%f", subtotal),
			Column3: fmt.Sprintf("%f", taxAmount),
			Column4: fmt.Sprintf("%f", shippingAmount),
			Column5: fmt.Sprintf("%f", discountAmount),
		})
		if err != nil {
			r.logger.Error("update order amounts failed", zap.Error(err))
			return fmt.Errorf("update order amounts: %w", err)
		}

		// Fields to parse from the database response
		fields := map[string]string{
			"subtotal":        row.Subtotal,
			"tax amount":      row.TaxAmount,
			"shipping amount": row.ShippingAmount,
			"discount amount": row.DiscountAmount,
			"vat amount":      row.VatAmount,
			"grand total":     row.GrandTotal,
		}

		parsedFields := make(map[string]float64, len(fields))
		for fieldName, fieldValue := range fields {
			parsedValue, err := strconv.ParseFloat(fieldValue, 64)
			if err != nil {
				r.logger.Error(fmt.Sprintf("failed to parse %s", fieldName), zap.Error(err))
				return fmt.Errorf("parse %s: %w", fieldName, err)
			}
			parsedFields[fieldName] = parsedValue
		}

		orderAmounts = model.OrderAmounts{
			SubTotal:       parsedFields["subtotal"],
			TaxAmount:      parsedFields["tax amount"],
			ShippingAmount: parsedFields["shipping amount"],
			DiscountAmount: parsedFields["discount amount"],
			VatAmount:      parsedFields["vat amount"],
			GrandTotal:     parsedFields["grand total"],
		}
		return nil
	})

	if err != nil {
		r.logger.Error("update order amounts transaction failed", zap.Error(err))
		return nil, fmt.Errorf("update order amounts transaction: %w", err)
	}

	return &orderAmounts, nil
}
