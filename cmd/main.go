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

	// Initialize services
	userService := services.NewUserService(userRepo, tokenRepo, &cfg)
	categoryService := services.NewCategoryService(categoryRepo, productColorRepo, logger)
	productService := services.NewProductService(
		productRepo,
		productVariantRepo,
		productImageRepo,
		productSpecificationRepo,
		categoryRepo,
		productColorRepo,
		productOptionRepo,
		logger,
		&cfg,
		s3Client,
	)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	productHandler := handlers.NewProductHandler(productService)
	productVariantHandler := handlers.NewProductVariantHandler(productService)
	productImageHandler := handlers.NewProductImageHandler(productService, s3Client, cfg.AWSBucketName)
	productSpecificationHandler := handlers.NewProductSpecificationHandler(productService)

	// Setup router
	r := setupRouter(logger, userHandler, categoryHandler, productHandler, productVariantHandler, productImageHandler, productSpecificationHandler)

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
) *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.CORS(logger))

	// Serve static files from the "uploads/profile" and "uploads/product-image" directories
	r.PathPrefix("/uploads/profile/").Handler(http.StripPrefix("/uploads/profile/", http.FileServer(http.Dir("uploads/profile"))))
	r.PathPrefix("/uploads/product-image/").Handler(http.StripPrefix("/uploads/product-image/", http.FileServer(http.Dir("uploads/product-image"))))

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
	categoryRouter.HandleFunc("/name/{name}", categoryHandler.GetCategoryByNameHandler).Methods(http.MethodOptions)
	categoryRouter.HandleFunc("/name/{name}", categoryHandler.GetCategoryByNameHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/parent/{parentId}/", categoryHandler.GetCategoriesByParentIDHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/products/count", categoryHandler.GetCategoriesWithProductsCountHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/tree", categoryHandler.GetCategoryTreeHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/hierarchy", categoryHandler.GetCategoryHierarchyHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/parent", categoryHandler.GetParentCategoriesHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/{id}/", categoryHandler.CheckCategoryExistenceHandler).Methods(http.MethodHead)
	categoryRouter.HandleFunc("/subcategories/count", categoryHandler.GetCategoriesWithSubcategoryCountHandler).Methods(http.MethodGet)

	// Product routes
	productRouter := r.PathPrefix("/api/products").Subrouter()
	productRouter.HandleFunc("", productHandler.CreateProductHandler).Methods(http.MethodPost)
	productRouter.HandleFunc("/{id}", productHandler.GetProductByIDHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("", productHandler.GetAllProductsHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/{id}", productHandler.UpdateProductHandler).Methods(http.MethodPut)
	productRouter.HandleFunc("/{id}", productHandler.SoftDeleteProductHandler).Methods(http.MethodDelete)
	productRouter.HandleFunc("/category/{id}", productHandler.GetProductsByCategoryIDHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/parent-category/{id}", productHandler.GetProductsByParentCategoryIDHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/parent-category/{id}", productHandler.GetProductsByParentCategoryIDHandler).Methods(http.MethodOptions)
	productRouter.HandleFunc("/filter/{category_id}", productHandler.GetProductsByFiltersHandler).Methods(http.MethodGet)

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

	return r
}
