package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
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

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize the logger
	logger, _ := zap.NewDevelopment()
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
	productColorRepo := repository.NewProductColourRepository(conn, logger)
	productOptionRepo := repository.NewProductOptionRepository(conn, logger)
	productSizeRepo := repository.NewProductSizeRepository(conn, logger)
	cartRepo := repository.NewCartRepository(conn, logger)
	orderRepo := repository.NewOrderRepository(conn, logger)
	guestCheckoutRepo := repository.NewGuestCheckoutRepository(conn, logger)
	orderItemRepo := repository.NewOrderItemRepository(conn, logger)
	paymentRepo := repository.NewPaymentRepository(conn, logger)
	discountRepo := repository.NewDiscountRepository(conn, logger)
	productAnalyticRepo := repository.NewProductAnalyticRepository(conn, logger)
	promotionRepo := repository.NewPromotionRepository(conn, logger)

	// Initialize services
	userService := services.NewUserService(userRepo, tokenRepo, &cfg)
	categoryService := services.NewCategoryService(categoryRepo, productColorRepo, logger, &cfg, s3Client)
	productService := services.NewProductService(
		productRepo,
		productVariantRepo,
		productImageRepo,
		productSpecificationRepo,
		categoryRepo,
		productColorRepo,
		productOptionRepo,
		productSizeRepo,
		discountRepo,
		logger,
		&cfg,
		s3Client,
	)
	productSizeService := services.NewProductSizeService(productSizeRepo, logger)
	cartService := services.NewCartService(logger, &cfg, cartRepo, productRepo, productImageRepo)
	orderService := services.NewOrderService(logger, guestCheckoutRepo, orderRepo, orderItemRepo, paymentRepo, userRepo, productRepo, &cfg)
	paymentService := services.NewPaymentService(paymentRepo, orderRepo, orderItemRepo, logger, &cfg)
	inquiryService := services.NewInquiryService(productRepo, logger, &cfg)
	productSEOService := services.NewProductSEOService(logger, &cfg, productRepo)
	productAnalyticService := services.NewProductAnalyticService(logger, &cfg, productAnalyticRepo, productImageRepo, discountRepo)
	promotionService := services.NewPromotionService(logger, &cfg, s3Client, promotionRepo, productRepo, productImageRepo, discountRepo)
	discountService := services.NewDiscountService(logger, discountRepo, productRepo)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	productHandler := handlers.NewProductHandler(productService, productSEOService)
	productVariantHandler := handlers.NewProductVariantHandler(productService)
	productImageHandler := handlers.NewProductImageHandler(productService)
	productSpecificationHandler := handlers.NewProductSpecificationHandler(productService)
	productOptionHandler := handlers.NewProductOptionHandler(productService)
	productColorHandler := handlers.NewProductColorHandler(productService)
	productSizeHandler := handlers.NewProductSizeHandler(productSizeService)
	cartHandler := handlers.NewCartHandler(cartService)
	orderHandler := handlers.NewOrderHandler(logger, orderService, paymentService)
	inquiryHandler := handlers.NewInquiryHandler(logger, inquiryService)
	productAnalyticHandler := handlers.NewProductAnalyticHandler(productAnalyticService)
	promotionHandler := handlers.NewPromotionHandler(promotionService)
	discountHandler := handlers.NewDiscountHandler(discountService)

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
		productColorHandler,
		productSizeHandler,
		cartHandler,
		orderHandler,
		inquiryHandler,
		productAnalyticHandler,
		promotionHandler,
		discountHandler,
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
	productColorHandler *handlers.ProductColorHandler,
	productSizeHandler *handlers.ProductSizeHandler,
	cartHandler *handlers.CartHandler,
	orderHandler *handlers.OrderHandler,
	inquiryHandler *handlers.InquiryHandler,
	productAnalyticHandler *handlers.ProductAnalyticHandler,
	promotionHandler *handlers.PromotionHandler,
	discountHandler *handlers.DiscountHandler,
) *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.CORS(logger))

	// Serve static files from the "uploads/profile" and "uploads/product-image" directories
	r.PathPrefix("/uploads/profile/").Handler(http.StripPrefix("/uploads/profile/", http.FileServer(http.Dir("uploads/profile"))))
	r.PathPrefix("/uploads/product-image/").Handler(http.StripPrefix("/uploads/product-image/", http.FileServer(http.Dir("uploads/product-image"))))

	// Inquiry routes
	inquiryRouter := r.PathPrefix("/api/inquiries").Subrouter()
	inquiryRouter.HandleFunc("", inquiryHandler.SubmitInquiry).Methods(http.MethodPost)

	// User routes
	resetPasswordRouter := r.PathPrefix("/reset-password").Subrouter()
	resetPasswordRouter.HandleFunc("", userHandler.ResetPassword).Methods(http.MethodGet)
	resetPasswordRouter.HandleFunc("", userHandler.ResetPassword).Methods(http.MethodPost)

	userRouter := r.PathPrefix("/api/users").Subrouter()
	userRouter.HandleFunc("", userHandler.ListUsers).Methods(http.MethodGet)
	userRouter.HandleFunc("/register", userHandler.RegisterUser).Methods(http.MethodPost)
	userRouter.HandleFunc("/login", userHandler.LoginUser).Methods(http.MethodPost)
	userRouter.HandleFunc("/refresh", userHandler.RefreshToken).Methods(http.MethodPost)

	protectedUserRouter := userRouter.PathPrefix("").Subrouter()
	protectedUserRouter.Use(middleware.Auth)
	protectedUserRouter.HandleFunc("/update-profile", userHandler.UpdateUserProfile).Methods(http.MethodPut)
	protectedUserRouter.HandleFunc("/reset-password", userHandler.RequestPasswordReset).Methods(http.MethodPut)

	// Category routes
	categoryRouter := r.PathPrefix("/api/categories").Subrouter()
	categoryRouter.HandleFunc("", categoryHandler.CreateCategoryHandler).Methods(http.MethodPost)
	categoryRouter.HandleFunc("/{id}/", categoryHandler.GetCategoryByIDHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("", categoryHandler.GetCategoriesHandler).Methods(http.MethodOptions)
	categoryRouter.HandleFunc("", categoryHandler.GetCategoriesHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/{id}/", categoryHandler.UpdateCategoryHandler).Methods(http.MethodPut)
	categoryRouter.HandleFunc("/{id}/", categoryHandler.SoftDeleteCategoryHandler).Methods(http.MethodDelete)
	categoryRouter.HandleFunc("/collections", categoryHandler.GetCollectionCategoriesHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/name/{name}", categoryHandler.GetCategoryByNameHandler).Methods(http.MethodOptions)
	categoryRouter.HandleFunc("/name/{name}", categoryHandler.GetCategoryByNameHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/parent/{parentId}", categoryHandler.GetCategoriesByParentIDHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/products/count", categoryHandler.GetCategoriesWithProductsCountHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/tree", categoryHandler.GetCategoryTreeHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/hierarchy", categoryHandler.GetCategoryHierarchyHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/parent", categoryHandler.GetParentCategoriesHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/{id}/", categoryHandler.CheckCategoryExistenceHandler).Methods(http.MethodHead)
	categoryRouter.HandleFunc("/subcategories/count", categoryHandler.GetCategoriesWithSubcategoryCountHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/upload-image", categoryHandler.UploadCategoryImageHandler).Methods(http.MethodPost)

	// Product routes
	productRouter := r.PathPrefix("/api/products").Subrouter()
	productRouter.HandleFunc("", productHandler.CreateProductHandler).Methods(http.MethodPost)
	productRouter.HandleFunc("/{slug}", productHandler.GetProductBySlugHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/{id}/", productHandler.GetProductByIDHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("", productHandler.GetAllProductsHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/{id}", productHandler.UpdateProductHandler).Methods(http.MethodPut)
	productRouter.HandleFunc("/{id}", productHandler.SoftDeleteProductHandler).Methods(http.MethodDelete)
	productRouter.HandleFunc("/{slug}/seo", productHandler.GetProductSEOHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/all/sitemap", productHandler.GetAllProductSitemapHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/actions/search", productHandler.SearchProductsHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/category/{id}", productHandler.GetProductsByCategoryIDHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/parent-category/{id}", productHandler.GetProductsByParentCategoryIDHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/parent-category/{id}", productHandler.GetProductsByParentCategoryIDHandler).Methods(http.MethodOptions)
	productRouter.HandleFunc("/filter/{category_id}", productHandler.GetProductsByFiltersHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/filter/all/options", productHandler.GetProductsByFilterOptionsHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/filter/options/{name}", productHandler.GetFilterOptionsByCategoryNameHandler).Methods(http.MethodGet)

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

	// Product Size routes
	productSizeRouter := r.PathPrefix("/api/product-sizes").Subrouter()
	productSizeRouter.HandleFunc("", productSizeHandler.CreateProductSizeHandler).Methods(http.MethodPost)
	productSizeRouter.HandleFunc("/{id}", productSizeHandler.UpdateProductSizeHandler).Methods(http.MethodPut)
	productSizeRouter.HandleFunc("/{id}", productSizeHandler.DeleteProductSizeHandler).Methods(http.MethodDelete)
	productSizeRouter.HandleFunc("/product/{id}", productSizeHandler.ListProductSizesByProductIDHandler).Methods(http.MethodGet)

	// Product Color routes
	productColorRouter := r.PathPrefix("/api/product-colors").Subrouter()
	productColorRouter.HandleFunc("", productColorHandler.CreateProductColorHandler).Methods(http.MethodPost)
	productColorRouter.HandleFunc("/{id}", productColorHandler.UpdateProductColorHandler).Methods(http.MethodPut)
	productColorRouter.HandleFunc("/{id}", productColorHandler.DeleteProductColorHandler).Methods(http.MethodDelete)
	productColorRouter.HandleFunc("/product/{id}", productColorHandler.ListProductColorsByProductIDHandler).Methods(http.MethodGet)

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

	// Product analytic routes
	productAnalyticRouter := r.PathPrefix("/api/product-analytics").Subrouter()
	productAnalyticRouter.HandleFunc("/best-sellers", productAnalyticHandler.GetBestSellerProducts).Methods(http.MethodGet)
	productAnalyticRouter.HandleFunc("/featured", productAnalyticHandler.GetFeaturedProducts).Methods(http.MethodGet)
	productAnalyticRouter.HandleFunc("/new-arrivals", productAnalyticHandler.GetNewArrivalProducts).Methods(http.MethodGet)
	productAnalyticRouter.HandleFunc("/daily-deals", productAnalyticHandler.GetDailyDealsProducts).Methods(http.MethodGet)

	// Promotion routes
	promotionRouter := r.PathPrefix("/api/promotions").Subrouter()
	promotionRouter.HandleFunc("", promotionHandler.CreatePromotion).Methods(http.MethodPost)
	promotionRouter.HandleFunc("", promotionHandler.GetPromotions).Methods(http.MethodGet)

	// Discount routes
	discountRouter := r.PathPrefix("/api/discounts").Subrouter()
	discountRouter.HandleFunc("", discountHandler.CreateDiscountHandler).Methods(http.MethodPost)

	return r
}
