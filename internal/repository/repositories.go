package repository

type Repositories struct {
	UserRepo                  *UserRepository
	TokenRepo                 *TokenRepository
	CategoryRepo              *CategoryRepository
	ProductRepo               *ProductRepository
	ProductVariantRepo        *ProductVariantRepository
	ProductImageRepo          *ProductImageRepository
	ProductSpecificationRepo  *ProductSpecificationRepository
	ProductOptionRepo         *ProductOptionRepository
	CartRepo                  CartRepository
	OrderRepo                 *OrderRepository
	GuestCheckoutRepo         *GuestCheckoutRepository
	OrderItemRepo             *OrderItemRepository
	PaymentRepo               *PaymentRepository
	DiscountRepo              *DiscountRepository
	ProductAnalyticRepo       *ProductAnalyticRepository
	PromotionRepo             *PromotionRepository
	RoleRepo                  *RoleRepository
	UserRoleRepo              *UserRoleRepository
	VerificationTokenRepo     VerificationTokenRepository
	PasswordResetRepo         PasswordResetRepository
	AdminRequestRepo          AdminRequestRepository
	ExchangeRateRepo          ExchangeRateRepository
	FilterCategoryProductRepo FilterCategoryProductRepository
	FilterProductRepo         FilterProductRepository
	ProductAttributeRepo      ProductAttributeRepository
	CompanyRepository         CompanyRepository
	SessionRepo               SessionRepository
}

type Dependencies struct {
	UserRepo                  *UserRepository
	TokenRepo                 *TokenRepository
	CategoryRepo              *CategoryRepository
	ProductRepo               *ProductRepository
	ProductVariantRepo        *ProductVariantRepository
	ProductImageRepo          *ProductImageRepository
	ProductSpecificationRepo  *ProductSpecificationRepository
	ProductOptionRepo         *ProductOptionRepository
	CartRepo                  CartRepository
	OrderRepo                 *OrderRepository
	GuestCheckoutRepo         *GuestCheckoutRepository
	OrderItemRepo             *OrderItemRepository
	PaymentRepo               *PaymentRepository
	DiscountRepo              *DiscountRepository
	ProductAnalyticRepo       *ProductAnalyticRepository
	PromotionRepo             *PromotionRepository
	RoleRepo                  *RoleRepository
	UserRoleRepo              *UserRoleRepository
	VerificationTokenRepo     VerificationTokenRepository
	PasswordResetRepo         PasswordResetRepository
	AdminRequestRepo          AdminRequestRepository
	ExchangeRateRepo          ExchangeRateRepository
	FilterCategoryProductRepo FilterCategoryProductRepository
	FilterProductRepo         FilterProductRepository
	ProductAttributeRepo      ProductAttributeRepository
	CompanyRepository         CompanyRepository
	SessionRepo               SessionRepository
}

// NewRepositories initializes and returns a Repositories instance with all repositories.
func NewRepositories(deps Dependencies) *Repositories {
	return &Repositories{
		UserRepo:                  deps.UserRepo,
		TokenRepo:                 deps.TokenRepo,
		CategoryRepo:              deps.CategoryRepo,
		ProductRepo:               deps.ProductRepo,
		ProductVariantRepo:        deps.ProductVariantRepo,
		ProductImageRepo:          deps.ProductImageRepo,
		ProductSpecificationRepo:  deps.ProductSpecificationRepo,
		ProductOptionRepo:         deps.ProductOptionRepo,
		CartRepo:                  deps.CartRepo,
		OrderRepo:                 deps.OrderRepo,
		GuestCheckoutRepo:         deps.GuestCheckoutRepo,
		OrderItemRepo:             deps.OrderItemRepo,
		PaymentRepo:               deps.PaymentRepo,
		DiscountRepo:              deps.DiscountRepo,
		ProductAnalyticRepo:       deps.ProductAnalyticRepo,
		PromotionRepo:             deps.PromotionRepo,
		RoleRepo:                  deps.RoleRepo,
		UserRoleRepo:              deps.UserRoleRepo,
		VerificationTokenRepo:     deps.VerificationTokenRepo,
		PasswordResetRepo:         deps.PasswordResetRepo,
		AdminRequestRepo:          deps.AdminRequestRepo,
		ExchangeRateRepo:          deps.ExchangeRateRepo,
		FilterCategoryProductRepo: deps.FilterCategoryProductRepo,
		FilterProductRepo:         deps.FilterProductRepo,
		ProductAttributeRepo:      deps.ProductAttributeRepo,
		CompanyRepository:         deps.CompanyRepository,
		SessionRepo:               deps.SessionRepo,
	}
}
