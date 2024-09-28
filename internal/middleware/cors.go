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

	// Precompute allowed origins map and allowAllOrigins flag
	originMap := make(map[string]struct{})
	allowAllOrigins := false
	if len(options.AllowedOrigins) == 1 && options.AllowedOrigins[0] == "*" {
		allowAllOrigins = true
	} else {
		for _, origin := range options.AllowedOrigins {
			originMap[origin] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log the incoming request
			logger.Info("Handling CORS", zap.String("method", r.Method), zap.String("path", r.URL.Path))

			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			if allowAllOrigins {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if _, ok := originMap[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				// Origin not allowed, return 403 Forbidden
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Set Vary header to indicate that the response varies based on the Origin header
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)

			if options.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Set allowed headers
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
