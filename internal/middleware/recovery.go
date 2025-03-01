// internal/middleware/recovery.go

package middleware

import (
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// Define a Prometheus counter for recovered panics
var (
	recoveredPanics = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_recovered_panics_total",
			Help: "Total number of panics recovered by the HTTP server.",
		},
	)
)

func init() {
	// Register the counter with Prometheus
	prometheus.MustRegister(recoveredPanics)
}

// RecoveryOptions defines the configuration options for RecoveryMiddleware.
type RecoveryOptions struct {
	// Logger is the zap.Logger instance used for logging.
	Logger *zap.Logger

	// EnableStackTrace determines whether to include stack traces in logs.
	EnableStackTrace bool

	// ResponseMessage is the message returned to the client upon recovery.
	// If empty, a default message is used.
	ResponseMessage string

	// AdditionalFields allows adding custom fields to the log entry.
	AdditionalFields []zap.Field
}

// RecoveryMiddleware returns a middleware that recovers from panics, logs the error, and responds with a 500 status code.
func RecoveryMiddleware(options RecoveryOptions) func(http.Handler) http.Handler {
	// Set default response message if not provided
	if options.ResponseMessage == "" {
		options.ResponseMessage = "Internal Server Error"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func(start time.Time) {
				if rec := recover(); rec != nil {
					// Increment the Prometheus counter
					recoveredPanics.Inc()

					// Capture the stack trace
					stackTrace := string(debug.Stack())

					// Prepare log fields
					logFields := []zap.Field{
						zap.Any("panic", rec),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.String("remote_addr", getClientIP(r)),
						zap.Duration("duration", time.Since(start)),
					}

					// Optionally include stack trace in logs
					if options.EnableStackTrace {
						logFields = append(logFields, zap.String("stack_trace", stackTrace))
					}

					// Append any additional custom fields
					if len(options.AdditionalFields) > 0 {
						logFields = append(logFields, options.AdditionalFields...)
					}

					// Log the panic
					options.Logger.Error("Panic recovered in HTTP handler", logFields...)

					// Respond to the client
					http.Error(w, options.ResponseMessage, http.StatusInternalServerError)
				}
			}(time.Now())

			// Proceed with the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP address from the request, considering proxies.
func getClientIP(r *http.Request) string {
	// Check the X-Forwarded-For header first
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, the first one is the client
		parts := splitAndTrim(xff, ",")
		if len(parts) > 0 {
			return parts[0]
		}
	}

	// Check the X-Real-IP header
	xrip := r.Header.Get("X-Real-Ip")
	if xrip != "" {
		return xrip
	}

	// Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // Fallback to the whole RemoteAddr
	}
	return ip
}

// splitAndTrim splits a string by the given separator and trims spaces from each part.
func splitAndTrim(s string, sep string) []string {
	parts := strings.Split(s, sep)
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}
