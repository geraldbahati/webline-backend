package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductImage struct {
	ID        uuid.UUID `json:"id"`
	ProductID string    `json:"product_id"`
	S3URL     string    `json:"s3_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Product struct {
	ID              uuid.UUID
	Name            string
	Description     string
	Price           string
	Stock           int32
	CategoryID      uuid.UUID
	IsActive        bool
	Featured        bool
	ImageURL        string
	DiscountPercent float64
}

type ProductDetail struct {
	ID              uuid.UUID
	Name            string
	Description     string
	Price           string
	Stock           int32
	CategoryID      uuid.UUID
	IsActive        bool
	Featured        bool
	DiscountPercent float64
	Images          []ProductImage
	Colors          []ProductColor
	Specifications  []ProductSpecification
	Options         []ProductOption
}

type ProductSchema struct {
	ID          uuid.UUID
	Name        string
	Description string
	Price       string
	Stock       int32
	CategoryID  uuid.UUID
	IsActive    bool
	Featured    bool
}

type ProductColor struct {
	ID        uuid.UUID
	ColorName string
}

type ProductSpecification struct {
	ID        uuid.UUID
	SpecName  string
	SpecValue string
}

type ProductVariant struct {
	ID           uuid.UUID
	VariantName  string
	VariantValue string
	Price        string
	Stock        int32
}

type ProductQueryResult struct {
	ID              uuid.UUID
	Name            string
	Price           string
	Stock           int32
	ImageURL        string
	DiscountPercent float64
}

type ProductOption struct {
	ID           uuid.UUID
	OptionName   string
	OptionValues []ProductOptionValue
}

type ProductOptionValue struct {
	ID              uuid.UUID
	ValueName       string
	AdditionalPrice float64
}

type ProductSize struct {
	ID              uuid.UUID
	ProductID       uuid.UUID
	Size            string
	AdditionalPrice string
}

type ProductCategoryFilterOption struct {
	Title         string
	Subcategories []ProductFilterOption
}

type ProductFilterOption struct {
	ID   uuid.UUID
	Name string
}

type ProductFilterOptions struct {
	Categories    []ProductCategoryFilterOption
	TotalProducts int64
}

type FilterProduct struct {
	ID              uuid.UUID
	Name            string
	Description     string
	Price           string
	Stock           int32
	CategoryID      uuid.UUID
	IsActive        bool
	Featured        bool
	ImageURL        string
	DiscountPercent float64
}
