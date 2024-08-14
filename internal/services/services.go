package services

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/repository"
)

type Services struct {
	User            *UserService
	Category        *CategoryService
	Product         *ProductService
	ProductSize     *ProductSizeService
	Cart            *CartService
	Order           *OrderService
	Payment         *PaymentService
	Inquiry         *InquiryService
	ProductSEO      *ProductSEOService
	ProductAnalytic *ProductAnalyticService
	Promotion       *PromotionService
	Discount        *DiscountService
	AdminRequest    *AdminRequestService
	Role            *RoleService
}

// NewServices initializes all services and returns a Services struct
func NewServices(repos *repository.Repositories, cfg appconfig.Config, logger *zap.Logger, s3Client *s3.Client) *Services {
	return &Services{
		User:            NewUserService(repos.User, repos.Role, repos.UserRole, repos.VerificationToken, repos.PasswordReset, repos.Token, &cfg, logger),
		Category:        NewCategoryService(repos.Category, repos.ProductColor, logger, &cfg, s3Client),
		Product:         NewProductService(repos.Product, repos.ProductVariant, repos.ProductImage, repos.ProductSpec, repos.Category, repos.ProductColor, repos.ProductOption, repos.ProductSize, repos.Discount, repos.User, repos.ExchangeRate, logger, &cfg, s3Client),
		ProductSize:     NewProductSizeService(repos.ProductSize, logger),
		Cart:            NewCartService(logger, &cfg, repos.Cart, repos.Product, repos.ProductImage),
		Order:           NewOrderService(logger, repos.GuestCheckout, repos.Order, repos.OrderItem, repos.Payment, repos.User, repos.Product, repos.ExchangeRate, &cfg),
		Payment:         NewPaymentService(repos.Payment, repos.Order, repos.OrderItem, logger, &cfg),
		Inquiry:         NewInquiryService(repos.Product, logger, &cfg),
		ProductSEO:      NewProductSEOService(logger, &cfg, repos.Product),
		ProductAnalytic: NewProductAnalyticService(logger, &cfg, repos.ProductAnalytic, repos.ProductImage, repos.Discount),
		Promotion:       NewPromotionService(logger, &cfg, s3Client, repos.Promotion, repos.Product, repos.ProductImage, repos.Discount, repos.User),
		Discount:        NewDiscountService(logger, repos.Discount, repos.Product),
		AdminRequest:    NewAdminRequestService(repos.AdminRequest, repos.User, logger, &cfg),
		Role:            NewRoleService(repos.Role, logger),
	}
}
