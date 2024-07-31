package app_errors

// Error codes
const (
	UseGoogleToLoginCode = 400
	EmailNotVerifiedCode = 403 // HTTP Status Code for Forbidden
	UserNotFoundCode     = 404
	TokenExpiredCode     = 401
	UnauthorizedCode     = 401
	AlreadyAdminCode     = 409
	InvalidTokenCode     = 401
)

// Factory functions for specific error cases
func NewEmailNotVerifiedError() *AppError {
	return NewAppError(EmailNotVerifiedCode, "Email is not verified", nil)
}

func NewUseGoogleToLoginError() *AppError {
	return NewAppError(UseGoogleToLoginCode, "Use Google to login", nil)
}

func NewUserNotFoundError() *AppError {
	return NewAppError(UserNotFoundCode, "User not found", nil)
}

func NewTokenExpiredError() *AppError {
	return NewAppError(TokenExpiredCode, "Token is expired", nil)
}

func NewUnauthorizedUserError() *AppError {
	return NewAppError(UnauthorizedCode, "User is not authorized", nil)
}

func NewAlreadyAdminError() *AppError {
	return NewAppError(AlreadyAdminCode, "User is already an admin", nil)
}

func NewInvalidTokenError() *AppError {
	return NewAppError(InvalidTokenCode, "Invalid token", nil)
}
