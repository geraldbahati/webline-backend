package handlers

import (
	"fmt"
	"net/http"
	"time"
	"weblineBackend/internal/middleware"
	"weblineBackend/internal/services"

	"go.uber.org/zap"
)

type SessionHandler struct {
	logger         *zap.Logger
	sessionService *services.SessionService
}

func NewSessionHandler(logger *zap.Logger, sessionService *services.SessionService) *SessionHandler {
	return &SessionHandler{
		logger:         logger,
		sessionService: sessionService,
	}
}

// CreateGuestSession creates a new guest session
func (h *SessionHandler) CreateGuestSession(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Creating new guest session")

	session, err := h.sessionService.CreateGuestSession(r.Context())
	if err != nil {
		h.logger.Error("Failed to create guest session", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to create guest session")
		return
	}

	h.logger.Info("Guest session created successfully",
		zap.String("sessionID", session.SessionID.String()))

	// Use helper functions so that cookie settings are correct for the current environment.
	secure := middleware.GetCookieSecureFlag()
	sameSite := middleware.GetCookieSameSitePolicy()

	http.SetCookie(w, &http.Cookie{
		Name:     "webline_session",
		Value:    session.SessionID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	h.logger.Info("Session cookie set",
		zap.String("sessionID", session.SessionID.String()),
		zap.String("secure", fmt.Sprintf("%t", secure)))

	http.SetCookie(w, &http.Cookie{
		Name:     "webline_csrf",
		Value:    session.CSRFToken,
		Path:     "/",
		HttpOnly: false, // CSRF token needs to be accessible by JavaScript
		Secure:   secure,
		SameSite: sameSite,
	})

	h.logger.Info("CSRF cookie set",
		zap.String("csrfToken", session.CSRFToken),
		zap.String("secure", fmt.Sprintf("%t", secure)))

	RespondWithJSON(w, http.StatusOK, session)
}

// MergeSession merges a guest session with a user session
func (h *SessionHandler) MergeSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.logger.Warn("Attempt to merge session without authentication")
		RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		h.logger.Warn("Attempt to merge session without session ID",
			zap.String("userID", userID.String()))
		RespondWithError(w, http.StatusBadRequest, "Session ID not provided")
		return
	}

	h.logger.Info("Attempting to merge sessions",
		zap.String("userID", userID.String()),
		zap.String("sessionID", sessionID))

	session, err := h.sessionService.MergeGuestSession(r.Context(), sessionID, userID)
	if err != nil {
		h.logger.Error("Failed to merge sessions",
			zap.Error(err),
			zap.String("userID", userID.String()),
			zap.String("sessionID", sessionID))
		RespondWithError(w, http.StatusInternalServerError, "Failed to merge sessions")
		return
	}

	h.logger.Info("Sessions merged successfully",
		zap.String("userID", userID.String()),
		zap.String("oldSessionID", sessionID),
		zap.String("newSessionID", session.SessionID.String()))

	secure := middleware.GetCookieSecureFlag()
	sameSite := middleware.GetCookieSameSitePolicy()

	// Update session cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "webline_session",
		Value:    session.SessionID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Expires:  time.Now().Add(30 * 24 * time.Hour), // persists for 30 days
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "webline_csrf",
		Value:    session.CSRFToken,
		Path:     "/",
		HttpOnly: false,
		Secure:   secure,
		SameSite: sameSite,
		Expires:  time.Now().Add(30 * 24 * time.Hour), // persists for 30 days
	})

	RespondWithJSON(w, http.StatusOK, session)
}
