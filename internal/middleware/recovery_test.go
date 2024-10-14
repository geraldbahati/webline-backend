// internal/middleware/recovery_test.go

package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestRecoveryMiddleware_WithErr(t *testing.T) {
	// Setup Zap logger with observer
	core, observed := observer.New(zapcore.ErrorLevel)
	logger := zap.New(core)

	// Define RecoveryMiddleware options
	recoveryOptions := RecoveryOptions{
		Logger:           logger,
		EnableStackTrace: true, // Enable stack traces in logs
		ResponseMessage:  "Internal Server Error",
		AdditionalFields: []zap.Field{
			zap.String("environment", "test"),
		},
	}

	// Define a handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Apply RecoveryMiddleware
	middleware := RecoveryMiddleware(recoveryOptions)(panicHandler)

	// Create a test server
	server := httptest.NewServer(middleware)
	defer server.Close()

	// Make a request to the panic handler
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	// Verify that the panic was logged
	logs := observed.All()
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	logEntry := logs[0]
	if logEntry.Level != zapcore.ErrorLevel {
		t.Errorf("Expected log level %s, got %s", zapcore.ErrorLevel, logEntry.Level)
	}

	if !strings.Contains(logEntry.Message, "Panic recovered in HTTP handler") {
		t.Errorf("Unexpected log message: %s", logEntry.Message)
	}

	// Check for the presence of the custom field
	found := false
	for _, field := range logEntry.Context {
		if field.Key == "environment" && field.String == "test" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected custom field 'environment=test' in log")
	}
}

func TestRecoveryMiddleware_WithoutErr(t *testing.T) {
	// Setup Zap logger with observer
	core, observed := observer.New(zapcore.ErrorLevel)
	logger := zap.New(core)

	// Define RecoveryMiddleware options
	recoveryOptions := RecoveryOptions{
		Logger:           logger,
		EnableStackTrace: false, // Disable stack traces in logs
		ResponseMessage:  "Internal Server Error",
		AdditionalFields: []zap.Field{
			zap.String("environment", "test"),
		},
	}

	// Define a handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Apply RecoveryMiddleware
	middleware := RecoveryMiddleware(recoveryOptions)(panicHandler)

	// Create a test server
	server := httptest.NewServer(middleware)
	defer server.Close()

	// Make a request to the panic handler
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	// Verify that the panic was logged
	logs := observed.All()
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	logEntry := logs[0]
	if logEntry.Level != zapcore.ErrorLevel {
		t.Errorf("Expected log level %s, got %s", zapcore.ErrorLevel, logEntry.Level)
	}

	if !strings.Contains(logEntry.Message, "Panic recovered in HTTP handler") {
		t.Errorf("Unexpected log message: %s", logEntry.Message)
	}

	// Check for the presence of the custom field
	found := false
	for _, field := range logEntry.Context {
		if field.Key == "environment" && field.String == "test" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected custom field 'environment=test' in log")
	}

	// Ensure stack_trace is not logged
	for _, field := range logEntry.Context {
		if field.Key == "stack_trace" {
			t.Error("Did not expect 'stack_trace' field in log")
		}
	}
}
