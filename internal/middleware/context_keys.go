package middleware

type contextKey string

const (
	SessionIDKey contextKey = "sessionID"
	UserIDKey    contextKey = "userID"
	SessionKey   contextKey = "session"
)
