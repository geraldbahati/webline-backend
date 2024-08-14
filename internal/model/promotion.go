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
