package handlers

type Handlers struct {
	UserHandler                 *UserHandler
	OrderHandler                *OrderHandler
	ProductHandler              *ProductHandler
	CartHandler                 *CartHandler
	RoleHandler                 *RoleHandler
	InquiryHandler              *InquiryHandler
	CategoryHandler             *CategoryHandler
	DiscountHandler             *DiscountHandler
	ProductImageHandler         *ProductImageHandler
	ProductOptionHandler        *ProductOptionHandler
	ProductVariantHandler       *ProductVariantHandler
	ProductAnalyticHandler      *ProductAnalyticHandler
	ProductSpecificationHandler *ProductSpecificationHandler
	PromotionHandler            *PromotionHandler
	GuestHandler                *GuestHandler
	SessionHandler              *SessionHandler
}

type Dependencies struct {
	UserHandler                 *UserHandler
	OrderHandler                *OrderHandler
	ProductHandler              *ProductHandler
	CartHandler                 *CartHandler
	RoleHandler                 *RoleHandler
	InquiryHandler              *InquiryHandler
	CategoryHandler             *CategoryHandler
	DiscountHandler             *DiscountHandler
	ProductImageHandler         *ProductImageHandler
	ProductOptionHandler        *ProductOptionHandler
	ProductVariantHandler       *ProductVariantHandler
	ProductAnalyticHandler      *ProductAnalyticHandler
	ProductSpecificationHandler *ProductSpecificationHandler
	PromotionHandler            *PromotionHandler
	GuestHandler                *GuestHandler
	SessionHandler              *SessionHandler
}

// NewHandlers initializes and returns a Handlers instance with all handlers.
func NewHandlers(deps Dependencies) *Handlers {
	return &Handlers{
		UserHandler:                 deps.UserHandler,
		OrderHandler:                deps.OrderHandler,
		ProductHandler:              deps.ProductHandler,
		CartHandler:                 deps.CartHandler,
		RoleHandler:                 deps.RoleHandler,
		InquiryHandler:              deps.InquiryHandler,
		CategoryHandler:             deps.CategoryHandler,
		DiscountHandler:             deps.DiscountHandler,
		ProductImageHandler:         deps.ProductImageHandler,
		ProductOptionHandler:        deps.ProductOptionHandler,
		ProductVariantHandler:       deps.ProductVariantHandler,
		ProductAnalyticHandler:      deps.ProductAnalyticHandler,
		ProductSpecificationHandler: deps.ProductSpecificationHandler,
		PromotionHandler:            deps.PromotionHandler,
		GuestHandler:                deps.GuestHandler,
		SessionHandler:              deps.SessionHandler,
	}
}
