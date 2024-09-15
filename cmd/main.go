package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"
	"weblineBackend/internal/repository/sqlc"

	"github.com/aws/aws-sdk-go-v2/config"
	"go.uber.org/zap/zapcore"

	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/handlers"
	"weblineBackend/internal/middleware"
	"weblineBackend/internal/repository"
	"weblineBackend/internal/services"
)

// CustomEncodeCaller creates a custom caller encoder that strips the /app prefix
func CustomEncodeCaller(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	// Convert the full file path to a string
	filePath := caller.String()

	// Remove the "/app" prefix
	filePath = strings.TrimPrefix(filePath, "/app/")

	// Add the modified file path to the encoder
	enc.AppendString(filePath)
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize the logger
	// Start with the development config
	zapConfig := zap.NewDevelopmentConfig()

	// Modify the encoder config to include the full file path in the caller
	zapConfig.EncoderConfig.EncodeCaller = CustomEncodeCaller

	// Build the logger using the modified config
	logger, _ := zapConfig.Build()
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			logger.Error("Failed to sync logger", zap.Error(err))
		}
	}(logger)

	logger.Info("Starting the application...")

	// Load configuration
	cfg := appconfig.LoadConfig()
	logger.Info("Configuration loaded successfully")

	// Initialize database connection
	conn, err := appconfig.NewDatabaseConnection(cfg.DbUrl)
	if err != nil {
		logger.Fatal("Failed to connect to the database", zap.Error(err))
	}
	defer func(conn *sql.DB) {
		err := conn.Close()
		if err != nil {
			logger.Error("Failed to close the database connection", zap.Error(err))
		}
	}(conn)
	logger.Info("Database connected successfully")

	// Initialize AWS S3 client
	logger.Info("Initializing AWS S3 client...")
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.AWSRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, ""),
		),
	)
	if err != nil {
		logger.Fatal("Failed to load AWS config", zap.Error(err))
	}
	s3Client := s3.NewFromConfig(awsConfig)
	logger.Info("AWS S3 client initialized successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(conn, logger)
	tokenRepo := repository.NewTokenRepository(conn, logger)
	categoryRepo := repository.NewCategoryRepository(conn, logger)
	productRepo := repository.NewProductRepository(conn, logger)
	productVariantRepo := repository.NewProductVariantRepository(conn, logger)
	productImageRepo := repository.NewProductImageRepository(conn, logger)
	productSpecificationRepo := repository.NewProductSpecificationRepository(conn, logger)

	productOptionRepo := repository.NewProductOptionRepository(conn, logger)

	cartRepo := repository.NewCartRepository(conn, logger)
	orderRepo := repository.NewOrderRepository(conn, logger)
	guestCheckoutRepo := repository.NewGuestCheckoutRepository(conn, logger)
	orderItemRepo := repository.NewOrderItemRepository(conn, logger)
	paymentRepo := repository.NewPaymentRepository(conn, logger)
	discountRepo := repository.NewDiscountRepository(conn, logger)
	productAnalyticRepo := repository.NewProductAnalyticRepository(conn, logger)
	promotionRepo := repository.NewPromotionRepository(conn, logger)
	roleRepo := repository.NewRoleRepository(conn, logger)
	userRoleRepo := repository.NewUserRoleRepository(conn, logger)
	verificationTokenRepoImpl := sqlc.NewVerificationTokenRepositoryImpl(conn, logger)
	passwordResetRepoImpl := sqlc.NewPasswordResetRepositoryImpl(conn, logger)
	adminRequestRepoImpl := sqlc.NewAdminRequestRepositoryImpl(conn, logger)
	exchangeRateRepoImpl := sqlc.NewExchangeRateRepositoryImpl(conn, logger)
	filterCategoryProductRepoImpl := sqlc.NewFilterCategoryProductRepoImpl(conn, logger)
	filterProductRepoImpl := sqlc.NewFilterProductRepoImpl(conn, logger)
	productAttributeRepoImpl := sqlc.NewProductAttributeRepoImpl(conn, logger)

	// Initialize services
	userService := services.NewUserService(userRepo, roleRepo, userRoleRepo, verificationTokenRepoImpl, passwordResetRepoImpl, adminRequestRepoImpl, tokenRepo, &cfg, logger, s3Client)
	categoryService := services.NewCategoryService(categoryRepo, userRepo, logger, &cfg, s3Client)
	productService := services.NewProductService(
		productRepo,
		productVariantRepo,
		productImageRepo,
		productSpecificationRepo,
		categoryRepo,
		productOptionRepo,
		discountRepo,
		userRepo,
		exchangeRateRepoImpl,
		logger,
		&cfg,
		s3Client,
	)
	cartService := services.NewCartService(logger, &cfg, cartRepo, productRepo, productImageRepo)
	orderService := services.NewOrderService(logger, guestCheckoutRepo, orderRepo, orderItemRepo, paymentRepo, userRepo, productRepo, exchangeRateRepoImpl, &cfg)
	paymentService := services.NewPaymentService(paymentRepo, orderRepo, orderItemRepo, logger, &cfg)
	inquiryService := services.NewInquiryService(productRepo, logger, &cfg)
	productSEOService := services.NewProductSEOService(logger, &cfg, productRepo)
	productAnalyticService := services.NewProductAnalyticService(logger, &cfg, productAnalyticRepo, productImageRepo, discountRepo)
	promotionService := services.NewPromotionService(logger, &cfg, s3Client, promotionRepo, productRepo, productImageRepo, discountRepo, userRepo)
	discountService := services.NewDiscountService(logger, discountRepo, productRepo)
	adminRequestService := services.NewAdminRequestService(adminRequestRepoImpl, userRepo, logger, &cfg)
	roleService := services.NewRoleService(roleRepo, logger)
	filterService := services.NewFilterService(logger, filterCategoryProductRepoImpl, filterProductRepoImpl, categoryRepo, &cfg)
	productAttributeService := services.NewProductAttributeService(productAttributeRepoImpl, logger)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService, adminRequestService, &cfg)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	productHandler := handlers.NewProductHandler(productService, productSEOService, filterService, productAttributeService)
	productVariantHandler := handlers.NewProductVariantHandler(productService)
	productImageHandler := handlers.NewProductImageHandler(productService)
	productSpecificationHandler := handlers.NewProductSpecificationHandler(productService)
	productOptionHandler := handlers.NewProductOptionHandler(productService)

	cartHandler := handlers.NewCartHandler(cartService)
	orderHandler := handlers.NewOrderHandler(logger, orderService, paymentService)
	inquiryHandler := handlers.NewInquiryHandler(logger, inquiryService)
	productAnalyticHandler := handlers.NewProductAnalyticHandler(productAnalyticService)
	promotionHandler := handlers.NewPromotionHandler(promotionService)
	discountHandler := handlers.NewDiscountHandler(discountService)
	roleHandler := handlers.NewRoleHandler(roleService)

	// Setup router
	r := setupRouter(
		logger,
		userHandler,
		categoryHandler,
		productHandler,
		productVariantHandler,
		productImageHandler,
		productSpecificationHandler,
		productOptionHandler,

		cartHandler,
		orderHandler,
		inquiryHandler,
		productAnalyticHandler,
		promotionHandler,
		discountHandler,
		roleHandler,
	)

	logger.Info("Router setup completed")

	// Start server
	serverAddress := ":" + cfg.Port
	logger.Info("Server listening", zap.String("port", cfg.Port))
	if err := http.ListenAndServe(serverAddress, r); err != nil {
		logger.Fatal("Server failed", zap.Error(err))
	}
}

func setupRouter(
	logger *zap.Logger,
	userHandler *handlers.UserHandler,
	categoryHandler *handlers.CategoryHandler,
	productHandler *handlers.ProductHandler,
	productVariantHandler *handlers.ProductVariantHandler,
	productImageHandler *handlers.ProductImageHandler,
	productSpecificationHandler *handlers.ProductSpecificationHandler,
	productOptionHandler *handlers.ProductOptionHandler,
	cartHandler *handlers.CartHandler,
	orderHandler *handlers.OrderHandler,
	inquiryHandler *handlers.InquiryHandler,
	productAnalyticHandler *handlers.ProductAnalyticHandler,
	promotionHandler *handlers.PromotionHandler,
	discountHandler *handlers.DiscountHandler,
	roleHandler *handlers.RoleHandler,
) *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.CORS(logger))
	r.Use(middleware.OptionalAuth(logger))

	// health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	// Serve static files from the "uploads/profile" and "uploads/product-image" directories
	r.PathPrefix("/uploads/profile/").Handler(http.StripPrefix("/uploads/profile/", http.FileServer(http.Dir("uploads/profile"))))
	r.PathPrefix("/uploads/product-image/").Handler(http.StripPrefix("/uploads/product-image/", http.FileServer(http.Dir("uploads/product-image"))))

	// verify email routes
	r.HandleFunc("/auth/verify-email", userHandler.VerifyEmail).Methods(http.MethodGet)

	// Inquiry routes
	inquiryRouter := r.PathPrefix("/api/inquiries").Subrouter()
	inquiryRouter.HandleFunc("", inquiryHandler.SubmitInquiry).Methods(http.MethodPost)

	// User routes
	userRouter := r.PathPrefix("/api/users").Subrouter()
	userRouter.HandleFunc("", userHandler.ListUsers).Methods(http.MethodGet)
	userRouter.HandleFunc("/{id}", userHandler.GetUserProfile).Methods(http.MethodGet)
	userRouter.HandleFunc("/register", userHandler.RegisterUser).Methods(http.MethodPost)
	userRouter.HandleFunc("/refresh", userHandler.RefreshToken).Methods(http.MethodPost)
	userRouter.HandleFunc("/login", userHandler.LoginUser).Methods(http.MethodPost)
	userRouter.HandleFunc("/reset-password", userHandler.ResetPassword).Methods(http.MethodPost)
	userRouter.HandleFunc("/reset-password/request", userHandler.RequestPasswordReset).Methods(http.MethodPost)
	userRouter.HandleFunc("/login/google", userHandler.LoginWithGoogle).Methods(http.MethodPost)
	userRouter.HandleFunc("/login/email-verified", userHandler.EmailVerified).Methods(http.MethodPost)
	userRouter.HandleFunc("/{id}/profile", userHandler.GetUserInfo).Methods(http.MethodGet)

	protectedUserRouter := userRouter.PathPrefix("").Subrouter()
	protectedUserRouter.Use(middleware.Auth(logger))
	protectedUserRouter.HandleFunc("/admin-requests", userHandler.RequestAdminRole).Methods(http.MethodPost)
	protectedUserRouter.HandleFunc("/approve", userHandler.ApproveAdminRole).Methods(http.MethodPost)
	protectedUserRouter.HandleFunc("/profile", userHandler.UpdateUserProfile).Methods(http.MethodPut)

	// Category routes
	categoryRouter := r.PathPrefix("/api/categories").Subrouter()
	categoryRouter.HandleFunc("/{id}/", categoryHandler.GetCategoryByIDHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("", categoryHandler.GetCategoriesHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/{id}/", categoryHandler.SoftDeleteCategoryHandler).Methods(http.MethodDelete)
	categoryRouter.HandleFunc("/collections", categoryHandler.GetCollectionCategoriesHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/name/{name}", categoryHandler.GetCategoryByNameHandler).Methods(http.MethodOptions)
	categoryRouter.HandleFunc("/name/{name}", categoryHandler.GetCategoryByNameHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/parent/{parentId}", categoryHandler.GetCategoriesByParentIDHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/products/count", categoryHandler.GetCategoriesWithProductsCountHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/tree", categoryHandler.GetCategoryTreeHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/hierarchy", categoryHandler.GetV2CategoryHierarchyHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/parent", categoryHandler.GetParentCategoriesHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/{id}/", categoryHandler.CheckCategoryExistenceHandler).Methods(http.MethodHead)
	categoryRouter.HandleFunc("/subcategories/count", categoryHandler.GetCategoriesWithSubcategoryCountHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/upload-image", categoryHandler.UploadCategoryImageHandler).Methods(http.MethodPost)

	// admin category
	adminCategoryRouter := r.PathPrefix("/api/v2/categories").Subrouter()
	adminCategoryRouter.HandleFunc("/{slug}/details", categoryHandler.GetCategoryDetailsHandler).Methods(http.MethodGet)
	adminCategoryRouter.HandleFunc("/hierarchy", categoryHandler.GetV2CategoryHierarchyHandler).Methods(http.MethodGet)
	adminCategoryRouter.HandleFunc("/{slug}/seo", categoryHandler.GetCategorySEOHandler).Methods(http.MethodGet)

	protectedAdminCategoryRouter := adminCategoryRouter.PathPrefix("").Subrouter()
	protectedAdminCategoryRouter.Use(middleware.Auth(logger))
	protectedAdminCategoryRouter.HandleFunc("", categoryHandler.CreateCategoryHandler).Methods(http.MethodPost)
	protectedAdminCategoryRouter.HandleFunc("/{id}", categoryHandler.DeleteCategoryHandler).Methods(http.MethodDelete)
	protectedAdminCategoryRouter.HandleFunc("/{id}", categoryHandler.SoftDeleteCategoryHandler).Methods(http.MethodPut)

	// Product routes
	productRouter := r.PathPrefix("/api/products").Subrouter()
	productRouter.HandleFunc("/{slug}", productHandler.GetProductBySlugHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/{slug}/images", productHandler.GetProductImagesBySlugHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/{slug}/pricing", productHandler.GetProductPricingBySlugHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/{slug}/specs", productHandler.GetProductSpecsBySlugHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("", productHandler.GetAllProductsHandler).Methods(http.MethodPost)
	productRouter.HandleFunc("/{slug}/seo", productHandler.GetProductSEOHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/all/sitemap", productHandler.GetAllProductSitemapHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/actions/search", productHandler.SearchProductsHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/category/{id}", productHandler.GetProductsByCategoryIDHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/filter/{category_id}", productHandler.GetFilteredCategoryProducts).Methods(http.MethodPost)
	productRouter.HandleFunc("/filter/all/options", productHandler.GetAllProductFilterOptionsHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/filter/options/{name}", productHandler.GetFilterOptionsByCategoryNameHandler).Methods(http.MethodGet)

	// admin product routes
	adminProductRouter := r.PathPrefix("/api/v2/products").Subrouter()
	adminProductRouter.HandleFunc("", productHandler.GetProductsHandler).Methods(http.MethodGet)
	adminProductRouter.HandleFunc("/{slug}/detail", productHandler.GetProductDetailHandler).Methods(http.MethodGet)
	adminProductRouter.HandleFunc("/meta-fields/{categoryID}", productHandler.GetProductMetaFieldsByCategoryIDHandler).Methods(http.MethodGet)

	protectedAdminProductRouter := adminProductRouter.PathPrefix("").Subrouter()
	protectedAdminProductRouter.Use(middleware.Auth(logger))
	protectedAdminProductRouter.HandleFunc("", productHandler.CreateV2ProductHandler).Methods(http.MethodPost)
	protectedAdminProductRouter.HandleFunc("", productHandler.DeleteProductsHandler).Methods(http.MethodDelete)
	protectedAdminProductRouter.HandleFunc("/{slug}", productHandler.DeleteProductHandler).Methods(http.MethodDelete)
	protectedAdminProductRouter.HandleFunc("/archive", productHandler.ArchiveProductsHandler).Methods(http.MethodPut)
	protectedAdminProductRouter.HandleFunc("/draft", productHandler.DraftProductsHandler).Methods(http.MethodPut)
	protectedAdminProductRouter.HandleFunc("/active", productHandler.ActivateProductsHandler).Methods(http.MethodPut)
	protectedAdminProductRouter.HandleFunc("/{slug}/archive", productHandler.ArchiveProductHandler).Methods(http.MethodPut)

	// Product Variant routes
	productVariantRouter := r.PathPrefix("/api/product-variants").Subrouter()
	productVariantRouter.HandleFunc("", productVariantHandler.CreateProductVariantHandler).Methods(http.MethodPost)
	productVariantRouter.HandleFunc("/{id}", productVariantHandler.GetProductVariantByIDHandler).Methods(http.MethodGet)
	productVariantRouter.HandleFunc("/product/{id}", productVariantHandler.ListProductVariantsByProductIDHandler).Methods(http.MethodGet)
	productVariantRouter.HandleFunc("/{id}", productVariantHandler.UpdateProductVariantHandler).Methods(http.MethodPut)
	productVariantRouter.HandleFunc("/{id}", productVariantHandler.DeleteProductVariantHandler).Methods(http.MethodDelete)

	// Product Image routes
	productImageRouter := r.PathPrefix("/api/product-images").Subrouter()
	productImageRouter.HandleFunc("", productImageHandler.CreateProductImageHandler).Methods(http.MethodPost)
	productImageRouter.HandleFunc("/{id}", productImageHandler.GetProductImageByIDHandler).Methods(http.MethodGet)
	productImageRouter.HandleFunc("/product/{product_id}", productImageHandler.GetProductImagesByProductIDHandler).Methods(http.MethodGet)
	productImageRouter.HandleFunc("/{id}", productImageHandler.UpdateProductImageHandler).Methods(http.MethodPut)
	productImageRouter.HandleFunc("/{id}", productImageHandler.DeleteProductImageHandler).Methods(http.MethodDelete)

	// Product Specification routes
	productSpecificationRouter := r.PathPrefix("/api/product-specifications").Subrouter()
	productSpecificationRouter.HandleFunc("", productSpecificationHandler.CreateProductSpecificationHandler).Methods(http.MethodPost)
	productSpecificationRouter.HandleFunc("", productSpecificationHandler.ListProductSpecificationsByProductIDHandler).Methods(http.MethodGet)
	productSpecificationRouter.HandleFunc("/{id}", productSpecificationHandler.DeleteProductSpecificationHandler).Methods(http.MethodDelete)

	// Product Option routes
	productOptionRouter := r.PathPrefix("/api/product-options").Subrouter()
	productOptionRouter.HandleFunc("/{id}", productOptionHandler.CreateProductOptionHandler).Methods(http.MethodPost)
	productOptionRouter.HandleFunc("/{id}", productOptionHandler.ListProductOptionsByProductIDHandler).Methods(http.MethodGet)
	productOptionRouter.HandleFunc("/{id}", productOptionHandler.DeleteProductOptionHandler).Methods(http.MethodDelete)
	productOptionRouter.HandleFunc("/{id}", productOptionHandler.UpdateProductOptionHandler).Methods(http.MethodPut)

	// Product Option Value routes
	productOptionValueRouter := r.PathPrefix("/api/product-option-values").Subrouter()
	productOptionValueRouter.HandleFunc("/{id}", productOptionHandler.CreateProductOptionValueHandler).Methods(http.MethodPost)
	productOptionValueRouter.HandleFunc("/{id}", productOptionHandler.ListProductOptionValuesByOptionIDHandler).Methods(http.MethodGet)
	productOptionValueRouter.HandleFunc("/{id}", productOptionHandler.DeleteProductOptionValueHandler).Methods(http.MethodDelete)
	productOptionValueRouter.HandleFunc("/{id}", productOptionHandler.UpdateProductOptionValueHandler).Methods(http.MethodPut)

	// Cart routes
	cartRouter := r.PathPrefix("/api/cart").Subrouter()
	cartRouter.HandleFunc("/create", cartHandler.CreateShoppingCartHandler).Methods(http.MethodPost)
	cartRouter.HandleFunc("/add/{cart_id}", cartHandler.AddToCartHandler).Methods(http.MethodPost)
	cartRouter.HandleFunc("/remove/{cart_id}", cartHandler.RemoveFromCartHandler).Methods(http.MethodDelete)
	cartRouter.HandleFunc("/items/{cart_id}", cartHandler.GetCartItemsHandler).Methods(http.MethodGet)
	cartRouter.HandleFunc("/clear/{cart_id}", cartHandler.ClearCartHandler).Methods(http.MethodPost)
	cartRouter.HandleFunc("/total/{cart_id}", cartHandler.CalculateCartTotalHandler).Methods(http.MethodGet)
	cartRouter.HandleFunc("/items/{cart_id}/update-quantity", cartHandler.UpdateCartItemQuantityHandler).Methods(http.MethodPut)
	cartRouter.HandleFunc("/user", cartHandler.GetShoppingCartByUserIDHandler).Methods(http.MethodPost)
	cartRouter.HandleFunc("/session", cartHandler.GetShoppingCartBySessionIDHandler).Methods(http.MethodPost)
	cartRouter.HandleFunc("/delete", cartHandler.DeleteShoppingCartHandler).Methods(http.MethodDelete)

	// Order routes
	orderRouter := r.PathPrefix("/api/orders").Subrouter()
	orderRouter.HandleFunc("", orderHandler.CreateOrder).Methods(http.MethodPost)
	orderRouter.HandleFunc("", orderHandler.ListOrders).Methods(http.MethodGet)
	orderRouter.HandleFunc("/{id}", orderHandler.GetOrder).Methods(http.MethodGet)
	orderRouter.HandleFunc("/pay", orderHandler.PayOrder).Methods(http.MethodPost)
	orderRouter.HandleFunc("/pay/status", orderHandler.GetPaymentStatus).Methods(http.MethodGet)
	orderRouter.HandleFunc("/pay/mpesa-callback", orderHandler.HandleMpesaCallback).Methods(http.MethodPost)
	orderRouter.HandleFunc("/{id}/cancel", orderHandler.CancelOrder).Methods(http.MethodPut)
	orderRouter.HandleFunc("/{id}/pay", orderHandler.ChangeOrderPaymentMethod).Methods(http.MethodPut)

	// Admin order routes
	adminOrderRouter := r.PathPrefix("/api/v2/orders").Subrouter()

	protectedAdminOrderRouter := adminOrderRouter.PathPrefix("").Subrouter()
	protectedAdminOrderRouter.Use(middleware.Auth(logger))
	protectedAdminOrderRouter.HandleFunc("/total-revenue", orderHandler.GetTotalRevenue).Methods(http.MethodGet)
	protectedAdminOrderRouter.HandleFunc("/monthly-sales", orderHandler.GetMonthlySales).Methods(http.MethodGet)
	protectedAdminOrderRouter.HandleFunc("/monthly-revenue", orderHandler.GetMonthlyRevenue).Methods(http.MethodGet)
	protectedAdminOrderRouter.HandleFunc("/sales-trend", orderHandler.GetSalesTrend).Methods(http.MethodGet)
	protectedAdminOrderRouter.HandleFunc("/recent-sales", orderHandler.GetRecentSales).Methods(http.MethodGet)
	protectedAdminOrderRouter.HandleFunc("/monthly-total-sales", orderHandler.GetTotalSalesCurrentMonth).Methods(http.MethodGet)
	protectedAdminOrderRouter.HandleFunc("/exchange-rate", orderHandler.GetExchangeRate).Methods(http.MethodGet)
	protectedAdminOrderRouter.HandleFunc("/exchange-rate", orderHandler.UpdateExchangeRate).Methods(http.MethodPut)

	// Product analytic routes
	productAnalyticRouter := r.PathPrefix("/api/product-analytics").Subrouter()
	productAnalyticRouter.HandleFunc("/best-sellers", productAnalyticHandler.GetBestSellerProducts).Methods(http.MethodGet)
	productAnalyticRouter.HandleFunc("/featured", productAnalyticHandler.GetFeaturedProducts).Methods(http.MethodGet)
	productAnalyticRouter.HandleFunc("/new-arrivals", productAnalyticHandler.GetNewArrivalProducts).Methods(http.MethodGet)
	productAnalyticRouter.HandleFunc("/daily-deals", productAnalyticHandler.GetDailyDealsProducts).Methods(http.MethodGet)

	// Promotion routes
	promotionRouter := r.PathPrefix("/api/promotions").Subrouter()
	promotionRouter.HandleFunc("", promotionHandler.GetPromotions).Methods(http.MethodGet)

	// Admin promotion routes
	adminPromotionRouter := r.PathPrefix("/api/v2/promotions").Subrouter()
	adminPromotionRouter.HandleFunc("", promotionHandler.GetV2Promotions).Methods(http.MethodGet)

	protectedAdminPromotionRouter := adminPromotionRouter.PathPrefix("").Subrouter()
	protectedAdminPromotionRouter.Use(middleware.Auth(logger))
	protectedAdminPromotionRouter.HandleFunc("", promotionHandler.CreateOrEditV2Promotion).Methods(http.MethodPost)
	protectedAdminPromotionRouter.HandleFunc("/{slug}", promotionHandler.GetPromotionDetails).Methods(http.MethodGet)

	// Discount routes
	discountRouter := r.PathPrefix("/api/discounts").Subrouter()

	protectedDiscountRouter := discountRouter.PathPrefix("").Subrouter()
	protectedDiscountRouter.Use(middleware.Auth(logger))
	protectedDiscountRouter.HandleFunc("", discountHandler.CreateDiscountHandler).Methods(http.MethodPost)

	// Role routes
	roleRouter := r.PathPrefix("/api/roles").Subrouter()
	roleRouter.HandleFunc("", roleHandler.CreateRole).Methods(http.MethodPost)

	return r
}
