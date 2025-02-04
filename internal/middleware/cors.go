package middleware

import (
	"net/http"
	"weblineBackend/internal/appconfig"

	"github.com/rs/cors"
	"go.uber.org/zap"
)

// CORS applies CORS settings using the rs/cors library.
// This middleware sets up allowed methods, headers and enables credentials (for cookie support).
func CORS(cfg *appconfig.Config, logger *zap.Logger) func(http.Handler) http.Handler {
	logger.Debug("Setting up CORS middleware",
		zap.Any("allowedOrigins", cfg.AllowedOrigins),
	)
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins, // e.g. []string{"https://yourdomain.com"}
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-CSRF-Token", "X-Session-ID"},
		AllowCredentials: true, // essential for cookies to be accepted in cross-origin requests
		Debug:            cfg.Env != "production",
	})
	return c.Handler
}
