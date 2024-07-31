package errors

const EmailNotVerifiedCode = 403 // HTTP Status Code for Forbidden

// EmailNotVerifiedError represents an error when a user's email is not verified.
type EmailNotVerifiedError struct {
	*AppError
}

// NewEmailNotVerifiedError creates a new EmailNotVerifiedError.
func NewEmailNotVerifiedError() *EmailNotVerifiedError {
	return &EmailNotVerifiedError{
		AppError: &AppError{
			Code:    EmailNotVerifiedCode,
			Message: "Email is not verified",
		},
	}
}

// NewUseGoogleToLogin
