// internal/middleware/rate_limit_test.go

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// mockRequestIDMiddleware simulates the RequestIDMiddleware for testing purposes.
func mockRequestIDMiddleware(reqID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// Assuming RequestIDKey is "requestID" as defined in your RequestIDMiddleware
			ctx = context.WithValue(ctx, "requestID", reqID)
			w.Header().Set("X-Request-ID", reqID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	// Setup Zap logger with observer
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	// Define a handler that returns 200 OK
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Initialize RateLimitMiddleware with 1 request per second, burst of 1, cleanup interval of 1 minute
	rps := 1.0
	burst := 1
	cleanupInterval := time.Minute
	rateLimitMiddleware := RateLimitMiddleware(rps, burst, cleanupInterval, logger) // Pass logger

	// Apply the middleware
	handler := rateLimitMiddleware(okHandler)

	// Create a test server
	server := httptest.NewServer(handler)
	defer server.Close()

	// Create a client with a fixed IP
	client := &http.Client{}
	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.RemoteAddr = "192.168.1.100:54321" // Fixed IP for testing

	// First request should pass
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("First request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Second request immediately should be rate limited
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Second request failed: %v", err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected status 429 Too Many Requests, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Wait for 1 second and try again
	time.Sleep(1 * time.Second)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Third request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK after wait, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify that logs were recorded
	logs := observed.All()
	expectedLogs := 2 // One for limiter creation, one for rate limit exceeded
	if len(logs) != expectedLogs {
		t.Fatalf("Expected %d log entries, got %d", expectedLogs, len(logs))
	}

	// Check specific log entries
	// Log 1: Created new rate limiter for IP
	log1 := logs[0]
	if log1.Message != "Created new rate limiter for IP" {
		t.Errorf("Unexpected log message: %s", log1.Message)
	}
	if log1.ContextMap()["ip"] != "192.168.1.100" {
		t.Errorf("Expected IP '192.168.1.100', got '%s'", log1.ContextMap()["ip"])
	}

	// Log 2: Rate limit exceeded
	log2 := logs[1]
	if log2.Message != "Rate limit exceeded" {
		t.Errorf("Unexpected log message: %s", log2.Message)
	}
	if log2.ContextMap()["ip"] != "192.168.1.100" {
		t.Errorf("Expected IP '192.168.1.100', got '%s'", log2.ContextMap()["ip"])
	}
	if log2.ContextMap()["path"] != "/test" {
		t.Errorf("Expected path '/test', got '%s'", log2.ContextMap()["path"])
	}
}

func TestRateLimitMiddleware_WithProxyHeaders(t *testing.T) {
	// Setup Zap logger with observer
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	// Define a handler that returns 200 OK
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Initialize RateLimitMiddleware with 2 requests per second, burst of 2, cleanup interval of 1 minute
	rps := 2.0
	burst := 2
	cleanupInterval := time.Minute
	rateLimitMiddleware := RateLimitMiddleware(rps, burst, cleanupInterval, logger) // Pass logger

	// Apply the middleware
	handler := rateLimitMiddleware(okHandler)

	// Create a test server
	server := httptest.NewServer(handler)
	defer server.Close()

	// Create a client with proxy headers
	client := &http.Client{}
	req, err := http.NewRequest("GET", server.URL+"/proxy", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("X-Forwarded-For", "203.0.113.195")
	req.Header.Set("X-Real-IP", "203.0.113.195")
	req.RemoteAddr = "192.168.1.100:54321" // This should be overridden by headers

	// First and second requests should pass
	for i := 0; i < 2; i++ {
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for request %d, got %d", i+1, resp.StatusCode)
		}
		resp.Body.Close()
	}

	// Third request immediately should be rate limited
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Third request failed: %v", err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected status 429 Too Many Requests for third request, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify that logs were recorded
	logs := observed.All()
	expectedLogs := 4 // One for limiter creation, two for HTTP Requests, one for rate limit exceeded
	if len(logs) != expectedLogs {
		t.Fatalf("Expected %d log entries, got %d", expectedLogs, len(logs))
	}

	// Check specific log entries
	// Log 1: Created new rate limiter for IP
	log1 := logs[0]
	if log1.Message != "Created new rate limiter for IP" {
		t.Errorf("Unexpected log message: %s", log1.Message)
	}
	if log1.ContextMap()["ip"] != "203.0.113.195" {
		t.Errorf("Expected IP '203.0.113.195', got '%s'", log1.ContextMap()["ip"])
	}

	// Logs 2 & 3: HTTP Requests
	for i := 1; i <= 2; i++ {
		log := logs[i]
		if log.Message != "HTTP Request" {
			t.Errorf("Unexpected log message: %s", log.Message)
		}
		if log.ContextMap()["method"] != "GET" {
			t.Errorf("Expected method 'GET', got '%s'", log.ContextMap()["method"])
		}
		if log.ContextMap()["path"] != "/proxy" {
			t.Errorf("Expected path '/proxy', got '%s'", log.ContextMap()["path"])
		}
		if log.ContextMap()["status"] != 200 {
			t.Errorf("Expected status 200, got '%d'", log.ContextMap()["status"])
		}
		if log.ContextMap()["bytes"] != 2 {
			t.Errorf("Expected bytes 2, got '%d'", log.ContextMap()["bytes"])
		}
		if log.ContextMap()["ip"] != "203.0.113.195" {
			t.Errorf("Expected IP '203.0.113.195', got '%s'", log.ContextMap()["ip"])
		}
		if durationMs, ok := log.ContextMap()["duration_ms"].(float64); !ok {
			t.Errorf("Expected duration_ms to be a float64, got %T", log.ContextMap()["duration_ms"])
		} else if durationMs <= 0 {
			t.Errorf("Expected duration_ms > 0, got '%v'", durationMs)
		}
	}

	// Log 4: Rate limit exceeded
	log4 := logs[3]
	if log4.Message != "Rate limit exceeded" {
		t.Errorf("Unexpected log message: %s", log4.Message)
	}
	if log4.ContextMap()["ip"] != "203.0.113.195" {
		t.Errorf("Expected IP '203.0.113.195', got '%s'", log4.ContextMap()["ip"])
	}
	if log4.ContextMap()["path"] != "/proxy" {
		t.Errorf("Expected path '/proxy', got '%s'", log4.ContextMap()["path"])
	}
}

func TestRateLimitMiddleware_WithLogging(t *testing.T) {
	// Setup Zap logger with observer
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	// Define a handler that returns 200 OK
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Initialize RateLimitMiddleware with 1 request per second, burst of 1, cleanup interval of 1 minute
	rps := 1.0
	burst := 1
	cleanupInterval := time.Minute
	rateLimitMiddleware := RateLimitMiddleware(rps, burst, cleanupInterval, logger) // Pass logger

	// Initialize LoggingMiddleware
	loggingMiddleware := LoggingMiddleware(logger)

	// Apply both RateLimitMiddleware and LoggingMiddleware
	handler := rateLimitMiddleware(
		loggingMiddleware(
			okHandler,
		),
	)

	// Create a test server
	server := httptest.NewServer(handler)
	defer server.Close()

	// Create a client with a fixed IP
	client := &http.Client{}
	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.RemoteAddr = "192.168.1.100:54321" // Fixed IP for testing

	// First request should pass
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("First request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Second request immediately should be rate limited
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Second request failed: %v", err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected status 429 Too Many Requests, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Wait for 1 second and try again
	time.Sleep(1 * time.Second)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Third request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK after wait, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify that logs were recorded
	logs := observed.All()
	expectedLogs := 4 // One for limiter creation, one for HTTP Request, one for rate limit exceeded
	if len(logs) != expectedLogs {
		t.Fatalf("Expected %d log entries, got %d", expectedLogs, len(logs))
	}

	// Check specific log entries
	// Log 1: Created new rate limiter for IP
	log1 := logs[0]
	if log1.Message != "Created new rate limiter for IP" {
		t.Errorf("Unexpected log message: %s", log1.Message)
	}
	if log1.ContextMap()["ip"] != "192.168.1.100" {
		t.Errorf("Expected IP '192.168.1.100', got '%s'", log1.ContextMap()["ip"])
	}

	// Log 2: HTTP Request (first request)
	log2 := logs[1]
	if log2.Message != "HTTP Request" {
		t.Errorf("Unexpected log message: %s", log2.Message)
	}
	if log2.ContextMap()["method"] != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", log2.ContextMap()["method"])
	}
	if log2.ContextMap()["path"] != "/test" {
		t.Errorf("Expected path '/test', got '%s'", log2.ContextMap()["path"])
	}
	if log2.ContextMap()["status"] != 200 {
		t.Errorf("Expected status 200, got '%d'", log2.ContextMap()["status"])
	}
	if log2.ContextMap()["bytes"] != 2 {
		t.Errorf("Expected bytes 2, got '%d'", log2.ContextMap()["bytes"])
	}
	if log2.ContextMap()["ip"] != "127.0.0.1" {
		t.Errorf("Expected IP '127.0.0.1', got '%s'", log2.ContextMap()["ip"])
	}
	if durationMs, ok := log2.ContextMap()["duration_ms"].(float64); !ok {
		t.Errorf("Expected duration_ms to be a float64, got %T", log2.ContextMap()["duration_ms"])
	} else if durationMs <= 0 {
		t.Errorf("Expected duration_ms > 0, got '%v'", durationMs)
	}

	// Log 3: Rate limit exceeded
	log3 := logs[2]
	if log3.Message != "Rate limit exceeded" {
		t.Errorf("Unexpected log message: %s", log3.Message)
	}
	if log3.ContextMap()["ip"] != "127.0.0.1" {
		t.Errorf("Expected IP '127.0.0.1', got '%s'", log3.ContextMap()["ip"])
	}
	if log3.ContextMap()["path"] != "/test" {
		t.Errorf("Expected path '/test', got '%s'", log3.ContextMap()["path"])
	}
}
