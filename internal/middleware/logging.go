// internal/middleware/logging.go

package middleware

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// statusRecorder is a wrapper for http.ResponseWriter that captures the status code and bytes written.
type statusRecorder struct {
	http.ResponseWriter
	status        int
	bytesWritten  int
	headerWritten bool
}

// newStatusRecorder initializes a new statusRecorder.
func newStatusRecorder(w http.ResponseWriter) *statusRecorder {
	// Default status is 200
	return &statusRecorder{ResponseWriter: w, status: http.StatusOK}
}

// WriteHeader captures the status code.
func (rec *statusRecorder) WriteHeader(code int) {
	if !rec.headerWritten {
		rec.status = code
		rec.ResponseWriter.WriteHeader(code)
		rec.headerWritten = true
	}
}

// Write captures the number of bytes written and ensures WriteHeader is called.
func (rec *statusRecorder) Write(b []byte) (int, error) {
	if !rec.headerWritten {
		// If WriteHeader hasn't been called yet, call it with 200
		rec.WriteHeader(http.StatusOK)
	}
	n, err := rec.ResponseWriter.Write(b)
	rec.bytesWritten += n
	return n, err
}

// LoggingMiddleware returns a middleware that logs each incoming HTTP request and its response status.
func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := newStatusRecorder(w)

			// Log that the request has started processing.
			logger.Debug("Started processing request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)

			// Proceed with the next handler
			next.ServeHTTP(rec, r)

			duration := time.Since(start)

			// Extract client IP considering proxies
			ip := getClientIP(r)

			// Retrieve Request ID from context if available
			reqID := getRequestID(r.Context())

			// Log the request details
			logger.Info("HTTP Request",
				zap.String("request_id", reqID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", rec.status),
				zap.Int("bytes", rec.bytesWritten),
				zap.Float64("duration_ms", float64(duration.Milliseconds())),
				zap.String("ip", ip),
				zap.String("user_agent", r.UserAgent()),
			)
		})
	}
}

// getRequestID retrieves the Request ID from the context, if available.
func getRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(requestIDContextKey).(string); ok {
		return reqID
	}
	return "unknown"
}
