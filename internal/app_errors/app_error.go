package app_errors

import "fmt"

// AppError represents a structured application error with an optional wrapped error.
type AppError struct {
	Code    int
	Message string
	Err     error
}

// Error implements the error interface for AppError.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%d: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error, if any.
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError.
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// NewNotFoundError creates a new AppError for a not found error.
func NewNotFoundError(message string) *AppError {
	return NewAppError(404, message, nil)
}

// NewInternalError creates a new AppError for an internal error.
func NewInternalError(message string, err error) *AppError {
	return NewAppError(500, message, err)
}

// NewValidationError creates a new AppError for a validation error.
func NewValidationError(message string) *AppError {
	return NewAppError(400, message, nil)
}

// NewAdminRequestPendingError creates a new AppError for an admin request pending error.
func NewAdminRequestPendingError() *AppError {
	return NewAppError(400, "Admin request is still pending", nil)
}

// NewSessionNotFoundError creates a new AppError for a session not found error.
func NewSessionNotFoundError() *AppError {
	return NewAppError(404, "Session not found", nil)
}
