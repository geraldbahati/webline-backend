package middleware

import (
	"net/http"
	"strings"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// set headers
		allowedOrigins := []string{"*"}
		origin := r.Header.Get("Origin")
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
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

		// Handle Content-Type
		if ct := r.Header.Get("Content-Type"); strings.Contains(ct, "multipart/form-data") {
			w.Header().Set("Content-Type", "multipart/form-data")
		} else {
			w.Header().Set("Content-Type", "application/json")
		}

		// handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// continue
		next.ServeHTTP(w, r)
	})
}
