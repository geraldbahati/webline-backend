// internal/middleware/logging_test.go

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggingMiddleware(t *testing.T) {
	// Setup Zap logger with observer
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	// Define a handler that returns a 200 OK
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Apply LoggingMiddleware
	logMiddleware := LoggingMiddleware(logger)(okHandler)

	// Create a test server
	req := httptest.NewRequest("GET", "http://example.com/api/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()

	// Serve the request
	logMiddleware.ServeHTTP(w, req)

	// Check the response
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify that the log was recorded
	logs := observed.All()
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	logEntry := logs[0]
	if logEntry.Message != "HTTP Request" {
		t.Errorf("Unexpected log message: %s", logEntry.Message)
	}

	// Check log fields
	fields := logEntry.ContextMap()
	if fields["method"] != "GET" {
		t.Errorf("Expected method GET, got %s", fields["method"])
	}
	if fields["path"] != "/api/test" {
		t.Errorf("Expected path /api/test, got %s", fields["path"])
	}
	// if fields["status"] != 200 {
	// 	t.Errorf("Expected status 200, got %d", fields["status"])
	// }
	// if fields["bytes"] != 2 {
	// 	t.Errorf("Expected bytes 2, got %d", fields["bytes"])
	// }
	if fields["ip"] != "192.168.1.100" {
		t.Errorf("Expected IP 192.168.1.100, got %s", fields["ip"])
	}
	if fields["user_agent"] != "" {
		t.Errorf("Expected empty user_agent, got %s", fields["user_agent"])
	}
	if _, exists := fields["request_id"]; !exists {
		t.Error("Expected request_id field in log")
	}
	if _, exists := fields["duration_ms"]; !exists {
		t.Error("Expected duration_ms field in log")
	}
}
