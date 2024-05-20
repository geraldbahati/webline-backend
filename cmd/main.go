package main

import (
	"go.uber.org/zap"
	"log"
	"net/http"
	"weblineBackend/internal/config"
	"weblineBackend/internal/handlers"
	"weblineBackend/internal/middleware"
	"weblineBackend/internal/repository"
	"weblineBackend/internal/services"
	"weblineBackend/pkg/logger"

	"github.com/gorilla/mux"
)

func main() {
	// initialize the logger
	logger.Init()
	logger.Info("Starting the application...")

	cfg := config.LoadConfig()

	// initialize database connection
	logger.Info("Connecting to database...")
	conn, err := config.NewDatabaseConnection(cfg.DbUrl)
	if err != nil {
		logger.Info("Error connecting to database: ", zap.Error(err))
	}
	logger.Info("Database connected successfully")

	// initialize repositories
	userRepo := repository.NewUserRepository(conn)

	// initialize services
	userService := services.NewUserService(userRepo)

	// initialize handlers
	userHandler := handlers.NewUserHandler(userService)

	// setup routes
	logger.Info("Initializing routers...")
	r := mux.NewRouter()
	r.Use(middleware.CORS)

	getUserRouter(r, userHandler)

	// start server
	log.Printf("Server listening on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}

func getUserRouter(r *mux.Router, userHandler *handlers.UserHandler) {
	//resetPasswordRouter := r.PathPrefix("/reset-password").Subrouter()
	//resetPasswordRouter.HandleFunc("", userHandler.ResetPassword).Methods(http.MethodGet)
	//resetPasswordRouter.HandleFunc("", userHandler.ResetPassword).Methods(http.MethodPost)

	userRouter := r.PathPrefix("/api/users").Subrouter()
	userRouter.HandleFunc("/register", userHandler.RegisterUser).Methods(http.MethodPost)
	//userRouter.HandleFunc("/login", userHandler.LoginUser).Methods(http.MethodPost)
	//userRouter.HandleFunc("/refresh", userHandler.RefreshToken).Methods(http.MethodPost)
	//
	//protectedUserRouter := userRouter.PathPrefix("").Subrouter()
	//protectedUserRouter.Use(middleware.Auth)
	//protectedUserRouter.HandleFunc("/update", userHandler.UpdateUser).Methods(http.MethodPut)
	//protectedUserRouter.HandleFunc("/update-profile-picture", userHandler.UpdateProfilePicture).Methods(http.MethodPut)
	//protectedUserRouter.HandleFunc("/reset-password", userHandler.RequestPasswordReset).Methods(http.MethodPut)
}
