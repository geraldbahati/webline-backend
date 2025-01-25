package middleware

type contextKey string

const (
	UserContextKey contextKey = "user"
	UserIDKey      contextKey = "userID"
	RequestIDKey   contextKey = "requestID"
)
