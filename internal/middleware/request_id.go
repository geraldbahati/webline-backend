// internal/middleware/request_id.go

package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestIDMiddleware assigns a unique ID to each request.
func RequestIDMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := uuid.New().String()
			ctx := context.WithValue(r.Context(), requestIDContextKey, reqID)
			w.Header().Set("X-Request-ID", reqID)

			// Add Request ID to logger
			logger = logger.With(zap.String("request_id", reqID))

			// Debug log to confirm request_id is set
			logger.Debug("Assigned Request ID to request", zap.String("request_id", reqID))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
