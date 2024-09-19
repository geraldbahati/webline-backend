package model

import (
	"github.com/google/uuid"
)

type DiscountSchema struct {
	ProductID       uuid.UUID `json:"productID"`
	DiscountPercentage float64   `json:"discountPercentage"`
}
