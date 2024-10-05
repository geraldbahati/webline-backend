package middleware

import (
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// CORSOptions holds configuration options for the CORS middleware
type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

// CORS middleware to handle CORS preflight requests and set appropriate headers
func CORS(logger *zap.Logger, options CORSOptions) func(http.Handler) http.Handler {
	allowedMethods := strings.Join(options.AllowedMethods, ", ")
	allowedHeaders := strings.Join(options.AllowedHeaders, ", ")

	// Precompute allowed origins map and check for wildcard origin
	allowedOrigins := make(map[string]struct{}, len(options.AllowedOrigins))
	allowAllOrigins := false
	for _, origin := range options.AllowedOrigins {
		if origin == "*" {
			allowAllOrigins = true
			if options.AllowCredentials {
				logger.Fatal("Cannot set AllowCredentials to true when AllowedOrigins contains '*'")
			}
		} else {
			allowedOrigins[origin] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			logger.Info("Received request", zap.String("Origin", origin), zap.String("Method", r.Method), zap.String("URL", r.URL.Path))

			// If Origin is present, handle CORS
			if origin != "" {
				// Always set Vary header
				w.Header().Set("Vary", "Origin")

				var allowedOrigin string
				if allowAllOrigins {
					allowedOrigin = "*"
				} else if _, ok := allowedOrigins[origin]; ok {
					allowedOrigin = origin
				} else {
					// Origin not allowed, return 403 Forbidden
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}

				// Set Access-Control-Allow-Origin
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)

				// Set Access-Control-Allow-Credentials if needed
				if options.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				// Set allowed methods and headers
				w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
				w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			}

			// Handle preflight request
			if r.Method == http.MethodOptions {
				logger.Info("Handled preflight CORS request", zap.String("path", r.URL.Path))
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Continue to the next handler
			next.ServeHTTP(w, r)
		})
	}
}
