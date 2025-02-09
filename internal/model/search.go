package model

import "github.com/google/uuid"

// SearchProductResult represents the basic information of a product returned as a result of a search query.
type SearchProductResult struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	PriceInKES float64   `json:"priceInKES"`
	Slug       string    `json:"slug"`
	Rank       float64   `json:"rank"`
}
