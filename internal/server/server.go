package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/handlers"
	"weblineBackend/internal/repository"
	"weblineBackend/internal/repository/sqlc"
	"weblineBackend/internal/routes"
	"weblineBackend/internal/services"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Server represents the application server with its dependencies.
type Server struct {
	Config   appconfig.Config
	Logger   *zap.Logger
	DB       *sql.DB
	S3Client *s3.Client
	Router   http.Handler
}

// NewServer initializes the server with configuration, logger, database, S3 client, repositories, services, and handlers.
// It returns a Server instance and an error if any step fails.
func NewServer(cfg appconfig.Config) (*Server, error) {
	logger := initLogger(cfg.Env)

	// 1. Establish Database Connection
	db, err := appconfig.NewDatabaseConnection(cfg.DbUrl)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 2. Initialize AWS S3 Client
	s3Client, err := initS3Client(cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize S3 client", zap.Error(err))
		return nil, fmt.Errorf("failed to initialize S3 client: %w", err)
	}

	// 3. Initialize Redis Client
	redisClient, err := initRedisClient(cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize Redis client", zap.Error(err))
		return nil, fmt.Errorf("failed to initialize Redis client: %w", err)
	}

	// 4. Initialize Repositories
	repos := initializeRepositories(db, logger)

	// 5. Initialize Services
	services := initializeServices(repos, cfg, logger, s3Client, redisClient)

	// 6. Initialize Handlers
	handlers := initializeHandlers(services, cfg, logger)

	// 7. Set Up Router
	router := routes.SetupRouter(cfg, logger, handlers, services.SessionService)

	// 8. Construct Server Instance
	server := &Server{
		Config:   cfg,
		Logger:   logger,
		DB:       db,
		S3Client: s3Client,
		Router:   router,
	}

	return server, nil
}

func initLogger(env string) *zap.Logger {
	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}
	cfg.EncoderConfig.EncodeCaller = customEncodeCaller
	logger, err := cfg.Build()
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

// initRedisClient initializes the Redis client.
func initRedisClient(cfg appconfig.Config, logger *zap.Logger) (*redis.Client, error) {
	logger.Info("Initializing Redis client...")
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdleConns,
	})

	// Ping Redis to verify the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to connect to Redis", zap.Error(err))
		return nil, err
	}

	logger.Info("Redis client initialized successfully")
	return rdb, nil
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
		CartRepo:                  sqlc.NewCartRepositoryImpl(db, logger),
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
		SessionRepo:               sqlc.NewSessionRepositoryImpl(db, logger),
	}
}

func initializeServices(repos *repository.Repositories, cfg appconfig.Config, logger *zap.Logger, s3Client *s3.Client, redisClient *redis.Client) *services.Services {
	cacheService := services.NewCacheService(redisClient, logger, cfg.RedisTTL, cfg.RedisRateLimit)
	sessionService := services.NewSessionService(logger, repos.SessionRepo, cacheService)
	orderService := services.NewOrderService(logger, repos.GuestCheckoutRepo, repos.OrderRepo, repos.OrderItemRepo, repos.PaymentRepo, repos.UserRepo, repos.ProductRepo, repos.DiscountRepo, repos.ExchangeRateRepo, repos.CompanyRepository, cacheService, &cfg, sessionService)

	return &services.Services{
		CacheService:            cacheService,
		UserService:             services.NewUserService(repos.UserRepo, repos.RoleRepo, repos.UserRoleRepo, repos.VerificationTokenRepo, repos.PasswordResetRepo, repos.AdminRequestRepo, repos.TokenRepo, repos.GuestCheckoutRepo, &cfg, logger, s3Client),
		CategoryService:         services.NewCategoryService(repos.CategoryRepo, repos.UserRepo, logger, &cfg, s3Client),
		ProductService:          services.NewProductService(repos.ProductRepo, repos.ProductVariantRepo, repos.ProductImageRepo, repos.ProductSpecificationRepo, repos.CategoryRepo, repos.ProductOptionRepo, repos.DiscountRepo, repos.UserRepo, repos.ExchangeRateRepo, cacheService, logger, &cfg, s3Client),
		CartService:             services.NewCartService(logger, &cfg, repos.CartRepo, repos.ProductRepo, repos.ProductImageRepo, repos.ExchangeRateRepo, cacheService, sessionService),
		OrderService:            orderService,
		PaymentService:          services.NewPaymentService(repos.PaymentRepo, repos.OrderRepo, repos.OrderItemRepo, logger, &cfg),
		InquiryService:          services.NewInquiryService(repos.ProductRepo, logger, &cfg),
		ProductSEOService:       services.NewProductSEOService(logger, &cfg, repos.ProductRepo, cacheService),
		ProductAnalyticService:  services.NewProductAnalyticService(logger, &cfg, repos.ProductAnalyticRepo, repos.ProductImageRepo, repos.DiscountRepo, cacheService),
		PromotionService:        services.NewPromotionService(logger, &cfg, s3Client, repos.PromotionRepo, repos.ProductRepo, repos.ProductImageRepo, repos.DiscountRepo, repos.UserRepo),
		DiscountService:         services.NewDiscountService(logger, repos.DiscountRepo, repos.ProductRepo),
		AdminRequestService:     services.NewAdminRequestService(repos.AdminRequestRepo, repos.UserRepo, logger, &cfg),
		RoleService:             services.NewRoleService(repos.RoleRepo, logger),
		FilterService:           services.NewFilterService(logger, repos.FilterCategoryProductRepo, repos.FilterProductRepo, repos.CategoryRepo, &cfg),
		ProductAttributeService: services.NewProductAttributeService(repos.ProductAttributeRepo, logger),
		SessionService:          services.NewSessionService(logger, repos.SessionRepo, cacheService),
		SearchService:           services.NewSearchService(repos.SearchRepo, cacheService, logger),
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
		CartHandler:                 handlers.NewCartHandler(svc.CartService, svc.SessionService, logger),
		OrderHandler:                handlers.NewOrderHandler(logger, svc.OrderService, svc.PaymentService, svc.UserService, svc.SessionService),
		InquiryHandler:              handlers.NewInquiryHandler(logger, svc.InquiryService),
		ProductAnalyticHandler:      handlers.NewProductAnalyticHandler(svc.ProductAnalyticService),
		PromotionHandler:            handlers.NewPromotionHandler(svc.PromotionService),
		DiscountHandler:             handlers.NewDiscountHandler(svc.DiscountService),
		RoleHandler:                 handlers.NewRoleHandler(svc.RoleService),
		GuestHandler:                handlers.NewGuestHandler(logger),
		SessionHandler:              handlers.NewSessionHandler(logger, svc.SessionService),
		SearchHandler:               handlers.NewSearchHandler(*svc.SearchService, logger),
	}
}
