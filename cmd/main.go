package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/server"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	// 1. Initialize Logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		// Ensure all buffered logs are flushed before exiting
		if err := logger.Sync(); err != nil {
			fmt.Printf("Failed to sync logger: %v\n", err)
		}
	}()

	// 2. Load Environment Variables
	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found or error loading .env file", zap.Error(err))
	}

	// 3. Load Configuration
	cfg, err := appconfig.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// 4. Create and Initialize the Server
	srv, err := server.NewServer(*cfg)
	if err != nil {
		logger.Fatal("Failed to create server", zap.Error(err))
	}

	// 5. Define the HTTP Server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", srv.Config.Port),
		Handler: srv.Router,
	}

	// 6. Channel to Listen for Server Errors
	serverErrors := make(chan error, 1)

	// 7. Start the Server in a Goroutine
	go func() {
		logger.Info("Starting server", zap.String("address", httpServer.Addr))
		serverErrors <- httpServer.ListenAndServe()
	}()

	// 8. Start the Metrics Server in a Goroutine
	go func() {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", promhttp.Handler())

		metricsServer := &http.Server{
			Addr:    ":2112", // Commonly used port for Prometheus metrics
			Handler: metricsMux,
		}

		logger.Info("Starting metrics server", zap.String("address", metricsServer.Addr))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Metrics server failed", zap.Error(err))
		}
	}()

	// 9. Channel to Listen for OS Interrupt Signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 10. Blocking Main Goroutine Until an Interrupt Signal is Received or Server Fails
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server encountered an error", zap.Error(err))
		}

	case sig := <-quit:
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
	}

	// 11. Initiate Graceful Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info("Initiating graceful shutdown")

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Fatal("Graceful shutdown failed", zap.Error(err))
	}

	logger.Info("Server gracefully stopped")
}
