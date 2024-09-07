package model

import (
	"github.com/google/uuid"
	"mime/multipart"
	"time"
)

type Promotion struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	ImageUrl    string `json:"imageUrl"`
	ProductSlug string `json:"productSlug"`
}

type PromotionSchema struct {
	ID          uuid.UUID
	Title       string
	Description string
	ImageUrl    string
	StartDate   time.Time
	EndDate     time.Time
	Slug        string
	Status      string
}

type V2Promotion struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Slug             string    `json:"slug"`
	Type             string    `json:"type"`
	ImageUrl         string    `json:"imageUrl"`
	NumberOfProducts int64     `json:"numberOfProducts"`
	Status           string    `json:"status"`
	StartDate        time.Time `json:"startDate"`
	EndDate          time.Time `json:"endDate"`
}

type CreatePromotionParams struct {
	Name         string
	Description  string
	Slug         string
	Status       string
	StartDate    time.Time
	EndDate      time.Time
	ProductSlugs []string
}

type ImageFile struct {
	File       multipart.File
	FileHeader *multipart.FileHeader
}

type PromotionDetails struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Slug        string             `json:"slug"`
	ImageUrl    string             `json:"imageUrl"`
	Status      string             `json:"status"`
	StartDate   time.Time          `json:"startDate"`
	EndDate     time.Time          `json:"endDate"`
	Products    []PromotionProduct `json:"products"`
}

type PromotionProduct struct {
	Slug     string  `json:"slug"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Discount float64 `json:"discount"`
	ImageURL string  `json:"imageURL"`
}
