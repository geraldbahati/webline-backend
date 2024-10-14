// internal/middleware/hsts.go

package middleware

import (
	"fmt"
	"net/http"
)

type HSTSOptions struct {
	MaxAge            int
	IncludeSubDomains bool
	Preload           bool
}

func HSTS(options HSTSOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only set HSTS for HTTPS requests
			if r.TLS != nil {
				hstsValue := fmt.Sprintf("max-age=%d", options.MaxAge)
				if options.IncludeSubDomains {
					hstsValue += "; includeSubDomains"
				}
				if options.Preload {
					hstsValue += "; preload"
				}
				w.Header().Set("Strict-Transport-Security", hstsValue)
			}
			next.ServeHTTP(w, r)
		})
	}
}
