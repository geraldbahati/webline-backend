// internal/middleware/metrics.go

package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// Define Prometheus metrics with optimized labels
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "handler", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets, // Customize if needed
		},
		[]string{"method", "handler", "status"},
	)
	// Gauge for tracking in-flight requests.
	inFlightRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "in_flight_requests",
			Help: "A gauge of in-flight requests being served by the handler.",
		},
	)
)

// Initialize registers the metrics with Prometheus
func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, inFlightRequests)
}

// MetricsMiddleware collects metrics for each HTTP request.
// It uses "handler" instead of "path" to reduce label cardinality.
// "handler" corresponds to the route's name in Gorilla Mux.
func MetricsMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Increment in-flight requests gauge.
			inFlightRequests.Inc()
			// Ensure gauge is decremented even if handler panics.
			defer inFlightRequests.Dec()

			start := time.Now()

			// Retrieve the current route
			route := mux.CurrentRoute(r)
			handlerName := "unknown"
			if route != nil {
				if name := route.GetName(); name != "" {
					handlerName = name
				} else {
					logger.Warn("Route without a name", zap.String("path", r.URL.Path))
				}
			} else {
				logger.Warn("No route found for request", zap.String("path", r.URL.Path))
			}

			// Use a ResponseWriter wrapper to capture the status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)
			duration := time.Since(start).Seconds()

			method := r.Method
			status := strconv.Itoa(rw.statusCode)

			// Increment Prometheus counters
			httpRequestsTotal.WithLabelValues(method, handlerName, status).Inc()
			httpRequestDuration.WithLabelValues(method, handlerName, status).Observe(duration)

			// Conditional Logging
			switch {
			case rw.statusCode >= 500:
				logger.Error("Handled request with server error",
					zap.String("method", method),
					zap.String("handler", handlerName),
					zap.String("status", status),
					zap.Float64("duration_seconds", duration),
				)
			case rw.statusCode >= 400:
				logger.Warn("Handled request with client error",
					zap.String("method", method),
					zap.String("handler", handlerName),
					zap.String("status", status),
					zap.Float64("duration_seconds", duration),
				)
			default:
				// For successful requests, log at Debug level to reduce log verbosity in production
				logger.Debug("Handled request",
					zap.String("method", method),
					zap.String("handler", handlerName),
					zap.String("status", status),
					zap.Float64("duration_seconds", duration),
				)
			}
		})
	}
}

// responseWriter is a wrapper to capture the status code and bytes written
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the status code if WriteHeader wasn't called
func (rw *responseWriter) Write(b []byte) (int, error) {
	// If WriteHeader hasn't been called yet, default to 200
	if rw.statusCode == http.StatusOK {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}
