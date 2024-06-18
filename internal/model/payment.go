package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderPayment struct {
	ID              uuid.UUID
	OrderID         uuid.UUID
	PaymentID       string
	Amount          string
	CreatedAt       time.Time
	PaymentMethodID int32
	PaymentStatusID int32
}

type OrderPaymentResponse struct {
	ID             uuid.UUID
	OrderNumber    string
	OrderCreatedAt time.Time
	CustomerName   string
	Amount         string
}
