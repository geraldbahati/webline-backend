package model

import (
	"github.com/google/uuid"
)

type Promotion struct {
	ProductID         uuid.UUID `json:"productID"`
	Tagline           string    `json:"tagline"`
	MainTitle         string    `json:"mainTitle"`
	SubTitle          string    `json:"subTitle"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	Discount          float64   `json:"discount"`
	PromotionImageUrl string    `json:"promotionImageUrl"`
	ProductImageUrl   string    `json:"productImageUrl"`
}

type PromotionSchema struct {
	ID          uuid.UUID
	Tagline     string
	MainTitle   string
	SubTitle    string
	Title       string
	Description string
	ImageUrl    string
}
