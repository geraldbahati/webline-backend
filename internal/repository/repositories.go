package repository

import (
	"database/sql"
	"go.uber.org/zap"
	"weblineBackend/internal/repository/sqlc"
)

type Repositories struct {
	User              *UserRepository
	Token             *TokenRepository
	Category          *CategoryRepository
	Product           *ProductRepository
	ProductVariant    *ProductVariantRepository
	ProductImage      *ProductImageRepository
	ProductSpec       *ProductSpecificationRepository
	ProductColor      *ProductColourRepository
	ProductOption     *ProductOptionRepository
	ProductSize       *ProductSizeRepository
	Cart              *CartRepository
	Order             *OrderRepository
	GuestCheckout     *GuestCheckoutRepository
	OrderItem         *OrderItemRepository
	Payment           *PaymentRepository
	Discount          *DiscountRepository
	ProductAnalytic   *ProductAnalyticRepository
	Promotion         *PromotionRepository
	Role              *RoleRepository
	UserRole          *UserRoleRepository
	VerificationToken *sqlc.VerificationTokenRepositoryImpl
	PasswordReset     *sqlc.PasswordResetRepositoryImpl
	AdminRequest      *sqlc.AdminRequestRepositoryImpl
	ExchangeRate      *sqlc.ExchangeRateRepoImpl
}

// NewRepositories initializes all repositories and returns a Repositories struct
func NewRepositories(conn *sql.DB, logger *zap.Logger) *Repositories {
	return &Repositories{
		User:              NewUserRepository(conn, logger),
		Token:             NewTokenRepository(conn, logger),
		Category:          NewCategoryRepository(conn, logger),
		Product:           NewProductRepository(conn, logger),
		ProductVariant:    NewProductVariantRepository(conn, logger),
		ProductImage:      NewProductImageRepository(conn, logger),
		ProductSpec:       NewProductSpecificationRepository(conn, logger),
		ProductColor:      NewProductColourRepository(conn, logger),
		ProductOption:     NewProductOptionRepository(conn, logger),
		ProductSize:       NewProductSizeRepository(conn, logger),
		Cart:              NewCartRepository(conn, logger),
		Order:             NewOrderRepository(conn, logger),
		GuestCheckout:     NewGuestCheckoutRepository(conn, logger),
		OrderItem:         NewOrderItemRepository(conn, logger),
		Payment:           NewPaymentRepository(conn, logger),
		Discount:          NewDiscountRepository(conn, logger),
		ProductAnalytic:   NewProductAnalyticRepository(conn, logger),
		Promotion:         NewPromotionRepository(conn, logger),
		Role:              NewRoleRepository(conn, logger),
		UserRole:          NewUserRoleRepository(conn, logger),
		VerificationToken: sqlc.NewVerificationTokenRepositoryImpl(conn, logger),
		PasswordReset:     sqlc.NewPasswordResetRepositoryImpl(conn, logger),
		AdminRequest:      sqlc.NewAdminRequestRepositoryImpl(conn, logger),
		ExchangeRate:      sqlc.NewExchangeRateRepositoryImpl(conn, logger),
	}
}
