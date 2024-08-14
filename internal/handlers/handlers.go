package handlers

import (
	"go.uber.org/zap"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/services"
)

type Handlers struct {
	User            *UserHandler
	Category        *CategoryHandler
	Product         *ProductHandler
	ProductVariant  *ProductVariantHandler
	ProductImage    *ProductImageHandler
	ProductSpec     *ProductSpecificationHandler
	ProductOption   *ProductOptionHandler
	ProductColor    *ProductColorHandler
	ProductSize     *ProductSizeHandler
	Cart            *CartHandler
	Order           *OrderHandler
	Inquiry         *InquiryHandler
	ProductAnalytic *ProductAnalyticHandler
	Promotion       *PromotionHandler
	Discount        *DiscountHandler
	Role            *RoleHandler
}

// NewHandlers initializes all handlers and returns a Handlers struct
func NewHandlers(services *services.Services, logger *zap.Logger, cfg appconfig.Config) *Handlers {
	return &Handlers{
		User:            NewUserHandler(services.User, services.AdminRequest, &cfg),
		Category:        NewCategoryHandler(services.Category),
		Product:         NewProductHandler(services.Product, services.ProductSEO),
		ProductVariant:  NewProductVariantHandler(services.Product),
		ProductImage:    NewProductImageHandler(services.Product),
		ProductSpec:     NewProductSpecificationHandler(services.Product),
		ProductOption:   NewProductOptionHandler(services.Product),
		ProductColor:    NewProductColorHandler(services.Product),
		ProductSize:     NewProductSizeHandler(services.ProductSize),
		Cart:            NewCartHandler(services.Cart),
		Order:           NewOrderHandler(logger, services.Order, services.Payment),
		Inquiry:         NewInquiryHandler(logger, services.Inquiry),
		ProductAnalytic: NewProductAnalyticHandler(services.ProductAnalytic),
		Promotion:       NewPromotionHandler(services.Promotion),
		Discount:        NewDiscountHandler(services.Discount),
		Role:            NewRoleHandler(services.Role),
	}
}
