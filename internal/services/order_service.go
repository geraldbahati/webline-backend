package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrderService struct {
	logger              *zap.Logger
	guestCheckoutRepo   *repository.GuestCheckoutRepository
	orderRepository     *repository.OrderRepository
	orderItemRepository *repository.OrderItemRepository
	userRepo            *repository.UserRepository
}

func NewOrderService(logger *zap.Logger, guestCheckoutRepo *repository.GuestCheckoutRepository, orderRepository *repository.OrderRepository, orderItemRepository *repository.OrderItemRepository, userRepo *repository.UserRepository) *OrderService {
	return &OrderService{
		logger:              logger,
		guestCheckoutRepo:   guestCheckoutRepo,
		orderRepository:     orderRepository,
		orderItemRepository: orderItemRepository,
		userRepo:            userRepo,
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

// CreateOrder creates a new order with items
func (s *OrderService) CreateOrder(ctx context.Context, orderParams *model.CreateOrderParams, items []model.CreateOrderItemParams) (*uuid.UUID, error) {
	userID := ctx.Value("userId")
	var userUUID uuid.NullUUID

	// Create or retrieve user or guest checkout
	if userID != nil {
		// UserID is already provided
		user, err := s.userRepo.GetUserByID(ctx, userID.(uuid.UUID))
		if err != nil {
			s.logger.Fatal("failed to get user", zap.Error(err))
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
		userUUID = uuid.NullUUID{UUID: user.ID, Valid: true}
	} else {
		// Check if the guest already exists
		existingGuest, err := s.guestCheckoutRepo.GetGuestCheckoutByEmail(ctx, orderParams.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			s.logger.Fatal("failed to check if guest exists", zap.Error(err))
			return nil, fmt.Errorf("failed to check if guest exists: %w", err)
		}

		var guestID uuid.UUID
		if existingGuest != nil {
			guestID = existingGuest.ID
		} else {
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
				s.logger.Fatal("failed to create guest checkout", zap.Error(err))
				return nil, fmt.Errorf("failed to create guest checkout: %w", err)
			}
			guestID = *newGuestID
		}

		userUUID = uuid.NullUUID{UUID: uuid.Nil, Valid: false}
		orderParams.GuestCheckoutID = uuid.NullUUID{UUID: guestID, Valid: true}
	}

	orderParam := &database.CreateOrderParams{
		UserID:          userUUID,
		GuestCheckoutID: orderParams.GuestCheckoutID,
		Total:           strconv.FormatFloat(orderParams.Total, 'f', -1, 64),
	}

	// Create order
	orderID, err := s.orderRepository.CreateOrder(ctx, orderParam)
	if err != nil {
		s.logger.Fatal("failed to create order", zap.Error(err))
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Create order items
	for _, item := range items {
		orderItem := &database.CreateOrderItemParams{
			OrderID:   uuid.NullUUID{UUID: *orderID, Valid: true},
			ProductID: uuid.NullUUID{UUID: item.ProductID, Valid: true},
			Quantity:  item.Quantity,
			Price:     item.Price,
		}

		orderItemID, err := s.orderItemRepository.CreateOrderItem(ctx, orderItem)
		if err != nil {
			s.logger.Fatal("failed to create order item", zap.Error(err))
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}

		for _, optionID := range item.ProductOptionIDs {
			if optionID.Valid {
				// Check if product option exists
				optionValue, err := s.orderItemRepository.GetProductOptionValueByID(ctx, optionID.UUID)
				if err != nil {
					s.logger.Fatal("failed to get product option", zap.Error(err))
					return nil, fmt.Errorf("failed to get product option: %w", err)
				}

				option, err := s.orderItemRepository.GetProductOptionByID(ctx, optionValue.OptionID.UUID)
				if err != nil {
					s.logger.Fatal("failed to get product option", zap.Error(err))
					return nil, fmt.Errorf("failed to get product option: %w", err)
				}

				orderItemOption := &database.CreateOrderItemOptionParams{
					OrderItemID:     uuid.NullUUID{UUID: *orderItemID, Valid: true},
					OptionType:      option.OptionName,
					OptionValue:     optionValue.ValueName,
					AdditionalPrice: optionValue.AdditionalPrice,
				}

				err = s.orderItemRepository.CreateOrderItemOption(ctx, orderItemOption)
				if err != nil {
					s.logger.Fatal("failed to create order item option", zap.Error(err))
					return nil, fmt.Errorf("failed to create order item option: %w", err)
				}
			}
		}

		if item.ColorID.Valid {
			// Check if color exists
			color, err := s.orderItemRepository.GetProductColorByID(ctx, item.ColorID.UUID)
			if err != nil {
				s.logger.Fatal("failed to get color", zap.Error(err))
				return nil, fmt.Errorf("failed to get color: %w", err)
			}

			orderItemOption := &database.CreateOrderItemOptionParams{
				OrderItemID:     uuid.NullUUID{UUID: *orderItemID, Valid: true},
				OptionType:      "Color",
				OptionValue:     color.ColorName,
				AdditionalPrice: sql.NullString{String: "0", Valid: true},
			}

			err = s.orderItemRepository.CreateOrderItemOption(ctx, orderItemOption)
			if err != nil {
				s.logger.Fatal("failed to create order item option", zap.Error(err))
				return nil, fmt.Errorf("failed to create order item option: %w", err)
			}
		}

		if item.SizeID.Valid {
			// Check if size exists
			size, err := s.orderItemRepository.GetProductSizeByID(ctx, item.SizeID.UUID)
			if err != nil {
				s.logger.Fatal("failed to get size", zap.Error(err))
				return nil, fmt.Errorf("failed to get size: %w", err)
			}

			orderItemOption := &database.CreateOrderItemOptionParams{
				OrderItemID:     uuid.NullUUID{UUID: *orderItemID, Valid: true},
				OptionType:      "Size",
				OptionValue:     size.Size,
				AdditionalPrice: size.AdditionalPrice,
			}

			err = s.orderItemRepository.CreateOrderItemOption(ctx, orderItemOption)
			if err != nil {
				s.logger.Fatal("failed to create order item option", zap.Error(err))
				return nil, fmt.Errorf("failed to create order item option: %w", err)
			}
		}
	}

	return orderID, nil
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
