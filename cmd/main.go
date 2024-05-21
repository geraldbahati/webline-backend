package main

import (
	"net/http"
	"weblineBackend/internal/config"
	"weblineBackend/internal/handlers"
	"weblineBackend/internal/middleware"
	"weblineBackend/internal/repository"
	"weblineBackend/internal/services"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {
	// Initialize the logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	logger.Info("Starting the application...")

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection
	conn, err := config.NewDatabaseConnection(cfg.DbUrl)
	if err != nil {
		logger.Fatal("Failed to connect to the database", zap.Error(err))
	}
	defer conn.Close()
	logger.Info("Database connected successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(conn, logger)
	tokenRepo := repository.NewTokenRepository(conn, logger)
	categoryRepo := repository.NewCategoryRepository(conn, logger)

	// Initialize services
	userService := services.NewUserService(userRepo, tokenRepo, &cfg)
	categoryService := services.NewCategoryService(categoryRepo, logger)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)

	// Setup router
	r := setupRouter(userHandler, categoryHandler)
	r.Use(middleware.CORS)

	// Start server
	serverAddress := ":" + cfg.Port
	logger.Info("Server listening", zap.String("port", cfg.Port))
	if err := http.ListenAndServe(serverAddress, r); err != nil {
		logger.Fatal("Server failed", zap.Error(err))
	}
}

func setupRouter(userHandler *handlers.UserHandler, categoryHandler *handlers.CategoryHandler) *mux.Router {
	r := mux.NewRouter()

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
	categoryRouter.HandleFunc("/{id}", categoryHandler.GetCategoryByIDHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("", categoryHandler.GetCategoriesHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/{id}", categoryHandler.UpdateCategoryHandler).Methods(http.MethodPut)
	categoryRouter.HandleFunc("/{id}", categoryHandler.SoftDeleteCategoryHandler).Methods(http.MethodDelete)
	categoryRouter.HandleFunc("/parent/{parentId}", categoryHandler.GetCategoriesByParentIDHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/products/count", categoryHandler.GetCategoriesWithProductsCountHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/tree", categoryHandler.GetCategoryTreeHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/{id}", categoryHandler.CheckCategoryExistenceHandler).Methods(http.MethodHead)
	categoryRouter.HandleFunc("/subcategories/count", categoryHandler.GetCategoriesWithSubcategoryCountHandler).Methods(http.MethodGet)

	return r
}
