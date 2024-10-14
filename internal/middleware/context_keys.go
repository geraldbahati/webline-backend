package middleware

type contextKey string

const (
	SessionIDKey   contextKey = "sessionID"
	UserContextKey contextKey = "user"
	UserIDKey      contextKey = "userID"
	SessionKey     contextKey = "session"
	RequestIDKey   contextKey = "requestID"
)
