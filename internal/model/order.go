package model

import (
	"time"

	"github.com/google/uuid"
)

type CreateGuestCheckoutParams struct {
	Email         string `json:"email"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Phone         string `json:"phone"`
	StreetAddress string `json:"street_address"`
	City          string `json:"city"`
	State         string `json:"state"`
	Country       string `json:"country"`
}

type CreateOrderItemParams struct {
	ProductID        uuid.UUID       `json:"product_id"`
	ProductOptionIDs []uuid.NullUUID `json:"product_option_id"`
	ColorID          uuid.NullUUID   `json:"color_id"`
	SizeID           uuid.NullUUID   `json:"size_id"`
	Quantity         int32           `json:"quantity"`
	Price            string          `json:"price"`
}

type CreateOrderParams struct {
	GuestCheckoutID uuid.NullUUID `json:"guest_checkout,omitempty"`
	Email           string        `json:"email"`
	FirstName       string        `json:"first_name"`
	LastName        string        `json:"last_name"`
	StreetAddress   string        `json:"street_address"`
	City            string        `json:"city"`
	State           string        `json:"state"`
	Country         string        `json:"country"`
	Phone           string        `json:"phone"`
	ShippingOption  string        `json:"shipping_option"`
	Total           float64       `json:"total"`
	ShippingMethod  string        `json:"shipping_method"`
	PaymentOption   string        `json:"payment_method"`
}

type OrderClientResponse struct {
	ID             uuid.UUID
	OrderNumber    string
	OrderCreatedAt time.Time
	CustomerName   string
	Phone          string
	Amount         float64
}
