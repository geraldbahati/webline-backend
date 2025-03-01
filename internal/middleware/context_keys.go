package middleware

// contextKey is a private type for keys stored in context.
type contextKey string

const (
	// userContextKey stores the authenticated user.
	userContextKey contextKey = "user"
	// userIDContextKey stores the user ID.
	userIDContextKey contextKey = "userID"
	// sessionContextKey stores the session instance.
	sessionContextKey contextKey = "session"
	// requestIDContextKey stores the request ID.
	requestIDContextKey contextKey = "requestID"
	// userTypeContextKey stores the user type.
	userTypeContextKey contextKey = "userType"
	// csrfTokenContextKey stores the CSRF token.
	csrfTokenContextKey contextKey = "csrfToken"
	// sessionIDContextKey stores the session ID.
	sessionIDContextKey contextKey = "sessionID"
)
