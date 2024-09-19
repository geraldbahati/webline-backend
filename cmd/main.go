package main

import (
	"fmt"
	"os"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/server"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Fatal("Error loading .env file", zap.Error(err))
	}

	// Load configuration
	cfg := appconfig.LoadConfig()

	// Create and start the server
	srv, err := server.NewServer(cfg)
	if err != nil {
		logger.Fatal("Failed to create server", zap.Error(err))
	}

	if err := srv.Run(); err != nil {
		logger.Fatal("Failed to run server", zap.Error(err))
	}
}
