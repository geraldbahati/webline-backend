package middleware

import (
	"go.uber.org/zap"
	"net/http"
)

// CORS middleware to handle CORS preflight requests and set appropriate headers
func CORS(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log the incoming request
			logger.Info("Handling CORS", zap.String("method", r.Method), zap.String("path", r.URL.Path))

			// Set allowed origins
			allowedOrigins := []string{"*"}
			origin := r.Header.Get("Origin")
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || origin == allowedOrigin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}

			// Set allowed methods
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")

			// Set allowed headers
			allowedHeaders := "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token"
			if requestedHeaders := r.Header.Get("Access-Control-Request-Headers"); requestedHeaders != "" {
				w.Header().Set("Access-Control-Allow-Headers", requestedHeaders)
			} else {
				w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			}

			// Handle preflight request
			if r.Method == http.MethodOptions {
				logger.Info("Handled preflight CORS request", zap.String("path", r.URL.Path))
				w.WriteHeader(http.StatusOK)
				return
			}

			// Continue to the next handler
			next.ServeHTTP(w, r)
		})
	}
}
