package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderPayment struct {
	ID                uuid.UUID `json:"id"`
	OrderID           uuid.UUID `json:"order_id"`
	CheckoutRequestID string    `json:"checkout_request_id"`
	Amount            string    `json:"amount"`
	CreatedAt         time.Time `json:"created_at"`
	PaymentMethodID   int32     `json:"payment_method_id"`
	PaymentStatusID   int32     `json:"payment_status_id"`
	ResultCode        int32     `json:"result_code"`
	ResultDesc        string    `json:"result_desc"`
}
