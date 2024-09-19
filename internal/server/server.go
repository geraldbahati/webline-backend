package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/handlers"
	"weblineBackend/internal/repository"
	"weblineBackend/internal/repository/sqlc"
	"weblineBackend/internal/routes"
	"weblineBackend/internal/services"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Server struct {
	config   appconfig.Config
	logger   *zap.Logger
	db       *sql.DB
	s3Client *s3.Client
	router   http.Handler
}

func NewServer(cfg appconfig.Config) (*Server, error) {
	logger := initLogger()
	db, err := appconfig.NewDatabaseConnection(cfg.DbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	s3Client, err := initS3Client(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize S3 client: %w", err)
	}

	repos := initializeRepositories(db, logger)
	services := initializeServices(repos, cfg, logger, s3Client)
	handlers := initializeHandlers(services, cfg, logger)

	router := routes.SetupRouter(logger, handlers)

	return &Server{
		config:   cfg,
		logger:   logger,
		db:       db,
		s3Client: s3Client,
		router:   router,
	}, nil
}

func (s *Server) Run() error {
	s.logger.Info("Server starting", zap.String("port", s.config.Port))
	return http.ListenAndServe(":"+s.config.Port, s.router)
}

func initLogger() *zap.Logger {
	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.EncoderConfig.EncodeCaller = customEncodeCaller
	logger, err := zapConfig.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	return logger
}

func customEncodeCaller(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(caller.TrimmedPath())
}

func initS3Client(cfg appconfig.Config, logger *zap.Logger) (*s3.Client, error) {
	logger.Info("Initializing AWS S3 client...")
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.AWSRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, err
	}
	logger.Info("AWS S3 client initialized successfully")
	return s3.NewFromConfig(awsCfg), nil
}

func initializeRepositories(db *sql.DB, logger *zap.Logger) *repository.Repositories {
	return &repository.Repositories{
		UserRepo:                  repository.NewUserRepository(db, logger),
		TokenRepo:                 repository.NewTokenRepository(db, logger),
		CategoryRepo:              repository.NewCategoryRepository(db, logger),
		ProductRepo:               repository.NewProductRepository(db, logger),
		ProductVariantRepo:        repository.NewProductVariantRepository(db, logger),
		ProductImageRepo:          repository.NewProductImageRepository(db, logger),
		ProductSpecificationRepo:  repository.NewProductSpecificationRepository(db, logger),
		ProductOptionRepo:         repository.NewProductOptionRepository(db, logger),
		CartRepo:                  repository.NewCartRepository(db, logger),
		OrderRepo:                 repository.NewOrderRepository(db, logger),
		GuestCheckoutRepo:         repository.NewGuestCheckoutRepository(db, logger),
		OrderItemRepo:             repository.NewOrderItemRepository(db, logger),
		PaymentRepo:               repository.NewPaymentRepository(db, logger),
		DiscountRepo:              repository.NewDiscountRepository(db, logger),
		ProductAnalyticRepo:       repository.NewProductAnalyticRepository(db, logger),
		PromotionRepo:             repository.NewPromotionRepository(db, logger),
		RoleRepo:                  repository.NewRoleRepository(db, logger),
		UserRoleRepo:              repository.NewUserRoleRepository(db, logger),
		VerificationTokenRepo:     sqlc.NewVerificationTokenRepositoryImpl(db, logger),
		PasswordResetRepo:         sqlc.NewPasswordResetRepositoryImpl(db, logger),
		AdminRequestRepo:          sqlc.NewAdminRequestRepositoryImpl(db, logger),
		ExchangeRateRepo:          sqlc.NewExchangeRateRepositoryImpl(db, logger),
		FilterCategoryProductRepo: sqlc.NewFilterCategoryProductRepoImpl(db, logger),
		FilterProductRepo:         sqlc.NewFilterProductRepoImpl(db, logger),
		ProductAttributeRepo:      sqlc.NewProductAttributeRepoImpl(db, logger),
		CompanyRepository:         sqlc.NewCompanyRepositoryImpl(db, logger),
	}
}

func initializeServices(repos *repository.Repositories, cfg appconfig.Config, logger *zap.Logger, s3Client *s3.Client) *services.Services {
	return &services.Services{
		UserService:             services.NewUserService(repos.UserRepo, repos.RoleRepo, repos.UserRoleRepo, repos.VerificationTokenRepo, repos.PasswordResetRepo, repos.AdminRequestRepo, repos.TokenRepo, repos.GuestCheckoutRepo, &cfg, logger, s3Client),
		CategoryService:         services.NewCategoryService(repos.CategoryRepo, repos.UserRepo, logger, &cfg, s3Client),
		ProductService:          services.NewProductService(repos.ProductRepo, repos.ProductVariantRepo, repos.ProductImageRepo, repos.ProductSpecificationRepo, repos.CategoryRepo, repos.ProductOptionRepo, repos.DiscountRepo, repos.UserRepo, repos.ExchangeRateRepo, logger, &cfg, s3Client),
		CartService:             services.NewCartService(logger, &cfg, repos.CartRepo, repos.ProductRepo, repos.ProductImageRepo),
		OrderService:            services.NewOrderService(logger, repos.GuestCheckoutRepo, repos.OrderRepo, repos.OrderItemRepo, repos.PaymentRepo, repos.UserRepo, repos.ProductRepo, repos.DiscountRepo, repos.ExchangeRateRepo, repos.CompanyRepository, &cfg),
		PaymentService:          services.NewPaymentService(repos.PaymentRepo, repos.OrderRepo, repos.OrderItemRepo, logger, &cfg),
		InquiryService:          services.NewInquiryService(repos.ProductRepo, logger, &cfg),
		ProductSEOService:       services.NewProductSEOService(logger, &cfg, repos.ProductRepo),
		ProductAnalyticService:  services.NewProductAnalyticService(logger, &cfg, repos.ProductAnalyticRepo, repos.ProductImageRepo, repos.DiscountRepo),
		PromotionService:        services.NewPromotionService(logger, &cfg, s3Client, repos.PromotionRepo, repos.ProductRepo, repos.ProductImageRepo, repos.DiscountRepo, repos.UserRepo),
		DiscountService:         services.NewDiscountService(logger, repos.DiscountRepo, repos.ProductRepo),
		AdminRequestService:     services.NewAdminRequestService(repos.AdminRequestRepo, repos.UserRepo, logger, &cfg),
		RoleService:             services.NewRoleService(repos.RoleRepo, logger),
		FilterService:           services.NewFilterService(logger, repos.FilterCategoryProductRepo, repos.FilterProductRepo, repos.CategoryRepo, &cfg),
		ProductAttributeService: services.NewProductAttributeService(repos.ProductAttributeRepo, logger),
	}
}

func initializeHandlers(svc *services.Services, cfg appconfig.Config, logger *zap.Logger) *handlers.Handlers {
	return &handlers.Handlers{
		UserHandler:                 handlers.NewUserHandler(svc.UserService, svc.AdminRequestService, &cfg),
		CategoryHandler:             handlers.NewCategoryHandler(svc.CategoryService),
		ProductHandler:              handlers.NewProductHandler(svc.ProductService, svc.ProductSEOService, svc.FilterService, svc.ProductAttributeService),
		ProductVariantHandler:       handlers.NewProductVariantHandler(svc.ProductService),
		ProductImageHandler:         handlers.NewProductImageHandler(svc.ProductService),
		ProductSpecificationHandler: handlers.NewProductSpecificationHandler(svc.ProductService),
		ProductOptionHandler:        handlers.NewProductOptionHandler(svc.ProductService),
		CartHandler:                 handlers.NewCartHandler(svc.CartService),
		OrderHandler:                handlers.NewOrderHandler(logger, svc.OrderService, svc.PaymentService, svc.UserService),
		InquiryHandler:              handlers.NewInquiryHandler(logger, svc.InquiryService),
		ProductAnalyticHandler:      handlers.NewProductAnalyticHandler(svc.ProductAnalyticService),
		PromotionHandler:            handlers.NewPromotionHandler(svc.PromotionService),
		DiscountHandler:             handlers.NewDiscountHandler(svc.DiscountService),
		RoleHandler:                 handlers.NewRoleHandler(svc.RoleService),
	}
}
