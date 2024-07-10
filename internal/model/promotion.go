package model

import (
	"github.com/google/uuid"
)

type Promotion struct {
	ProductID         uuid.UUID `json:"productID"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	Discount          float64   `json:"discount"`
	PromotionImageUrl string    `json:"promotionImageUrl"`
	ProductImageUrl   string    `json:"productImageUrl"`
}

type PromotionSchema struct {
	ID          uuid.UUID
	Title       string
	Description string
	ImageUrl    string
}
