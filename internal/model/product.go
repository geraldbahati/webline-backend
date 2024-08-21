package model

import (
	"encoding/json"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type ProductImage struct {
	ID        uuid.UUID `json:"id"`
	ProductID string    `json:"productID"`
	S3URL     string    `json:"s3URL"`
}

type Product struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Price           string    `json:"price"`
	ImageURL        string    `json:"imageURL"`
	DiscountPercent float64   `json:"discountPercent"`
	Slug            string    `json:"slug"`
}

type V2Product struct {
	Name         string    `json:"name"`
	Price        string    `json:"price"`
	Status       string    `json:"status"`
	ImageURL     string    `json:"imageURL"`
	CategoryName string    `json:"categoryName"`
	Discount     float64   `json:"discount"`
	Slug         string    `json:"slug"`
	CreatedAt    time.Time `json:"createdAt"`
	InPromotion  bool      `json:"inPromotion"`
	TotalSales   int32     `json:"totalSales"`
	PartNumber   string    `json:"partNumber"`
}

type V2ProductDetail struct {
	Name              string             `json:"name"`
	Description       string             `json:"description"`
	Price             string             `json:"price"`
	Stock             int32              `json:"stock"`
	PartNumber        string             `json:"partNumber"`
	Slug              string             `json:"slug"`
	MetaTitle         string             `json:"metaTitle"`
	MetaDescription   string             `json:"metaDescription"`
	MetaKeywords      string             `json:"metaKeywords"`
	ExchangeRate      float64            `json:"exchangeRate"`
	PricePerUnit      string             `json:"pricePerUnit"`
	ProfitMargin      string             `json:"profitMargin"`
	ParentCategoryID  uuid.UUID          `json:"parentCategoryID"`
	ProductMetafields []ProductMetafield `json:"productMetafields"`
	CategoryID        uuid.UUID          `json:"categoryID"`
	Status            string             `json:"status"`
	Specifications    json.RawMessage    `json:"specifications"`
	Images            json.RawMessage    `json:"images"`
}

type ProductMetafield struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type V2ProductSpec struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type V2ProductImage struct {
	Url string `json:"url"`
}

type ProductDetail struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type CreateProductRequest struct {
	Slug            string                  `json:"slug"`
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	Price           float64                 `json:"price"`
	Stock           int                     `json:"stock"`
	CategoryID      uuid.UUID               `json:"categoryID"`
	Status          string                  `json:"status"`
	PartNumber      string                  `json:"partNumber"`
	MetaTitle       string                  `json:"metaTitle"`
	MetaDescription string                  `json:"metaDescription"`
	MetaKeywords    string                  `json:"metaKeywords"`
	Images          []*multipart.FileHeader `json:"images"`
	Specifications  []Specification         `json:"specifications"`
	ImageUrls       []string                `json:"imageUrls"`
}

type Specification struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ProductSchema struct {
	ID           uuid.UUID
	Name         string
	Description  string
	USD          string
	Stock        int32
	CategoryID   uuid.UUID
	CategoryName string
	Status       string
	Featured     bool
	Slug         string
}

type ProductColor struct {
	ID        uuid.UUID `json:"id"`
	ColorName string    `json:"colorName"`
}

type ProductSpecification struct {
	ID        uuid.UUID `json:"id"`
	SpecName  string    `json:"specName"`
	SpecValue string    `json:"specValue"`
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
	Slug            string
}

type ProductOption struct {
	ID           uuid.UUID            `json:"id"`
	OptionName   string               `json:"optionName"`
	OptionValues []ProductOptionValue `json:"optionValues"`
}

type ProductOptionValue struct {
	ID              uuid.UUID `json:"id"`
	ValueName       string    `json:"valueName"`
	AdditionalPrice float64   `json:"additionalPrice"`
}

type ProductSize struct {
	ID              uuid.UUID
	ProductID       uuid.UUID
	Size            string
	AdditionalPrice string
}

type ProductCategoryFilterOption struct {
	Title         string                `json:"title"`
	Subcategories []ProductFilterOption `json:"subcategories"`
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
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Price           string    `json:"price"`
	Stock           int32     `json:"stock"`
	CategoryID      uuid.UUID `json:"categoryID"`
	IsActive        bool      `json:"isActive"`
	Featured        bool      `json:"featured"`
	ImageURL        string    `json:"imageURL"`
	DiscountPercent float64   `json:"discountPercent"`
	Slug            string    `json:"slug"`
}

// UnifiedParams defines the common parameters for product filtering.
type UnifiedParams struct {
	ID             uuid.UUID
	CategoryNames  []string
	ColorNames     []string
	ProcessorNames []string
	StorageNames   []string
	Sizes          []string
	PriceFrom      float64
	PriceTo        float64
	SortOrder      string
}

type ProductSitemap struct {
	ID        uuid.UUID
	UpdatedAt time.Time
}

type ProductSEO struct {
	ID          uuid.UUID `json:"id"`
	PartNumber  string    `json:"partNumber"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Keywords    string    `json:"keywords"`
	Price       string    `json:"price"`
	Brand       string    `json:"brand"`
	ImageUrl    string    `json:"imageUrl"`
}

type ProductPricing struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Price           string    `json:"price"`
	DiscountPercent float64   `json:"discountPercent"`
	ImageUrl        string    `json:"imageUrl"`
}

type ProductSpecs struct {
	Description string        `json:"description"`
	Specs       []ProductSpec `json:"specs"`
}

type ProductSpec struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Attribute struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}
