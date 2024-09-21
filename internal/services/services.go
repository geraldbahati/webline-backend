package services

type Services struct {
	UserService             *UserService
	OrderService            *OrderService
	ProductService          *ProductService
	CartService             *CartService
	RoleService             *RoleService
	InquiryService          *InquiryService
	CategoryService         *CategoryService
	DiscountService         *DiscountService
	ProductAnalyticService  *ProductAnalyticService
	PaymentService          *PaymentService
	FilterService           *FilterService
	ProductSEOService       *ProductSEOService
	AdminRequestService     *AdminRequestService
	ProductAttributeService *ProductAttributeService
	PromotionService        *PromotionService
	CacheService            CacheService
}

type Dependencies struct {
	UserService             *UserService
	OrderService            *OrderService
	ProductService          *ProductService
	CartService             *CartService
	RoleService             *RoleService
	InquiryService          *InquiryService
	CategoryService         *CategoryService
	DiscountService         *DiscountService
	ProductAnalyticService  *ProductAnalyticService
	PaymentService          *PaymentService
	FilterService           *FilterService
	ProductSEOService       *ProductSEOService
	AdminRequestService     *AdminRequestService
	ProductAttributeService *ProductAttributeService
	PromotionService        *PromotionService
	CacheService            CacheService
}

// NewServices initializes and returns a Services instance with all services.
func NewServices(deps Dependencies) *Services {
	return &Services{
		UserService:             deps.UserService,
		OrderService:            deps.OrderService,
		ProductService:          deps.ProductService,
		CartService:             deps.CartService,
		RoleService:             deps.RoleService,
		InquiryService:          deps.InquiryService,
		CategoryService:         deps.CategoryService,
		DiscountService:         deps.DiscountService,
		ProductAnalyticService:  deps.ProductAnalyticService,
		PaymentService:          deps.PaymentService,
		FilterService:           deps.FilterService,
		ProductSEOService:       deps.ProductSEOService,
		AdminRequestService:     deps.AdminRequestService,
		ProductAttributeService: deps.ProductAttributeService,
		PromotionService:        deps.PromotionService,
		CacheService:            deps.CacheService,
	}
}
