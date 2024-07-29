package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"weblineBackend/internal/model"
	"weblineBackend/internal/services"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// RegisterUser registers a new user
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var registerUserParams model.RegisterUserParams

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&registerUserParams); err != nil {
		RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Failed to decode request body")})
		return
	}

	// Create user
	user, err := h.userService.CreateUser(r.Context(), registerUserParams)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create user: %v", err))
		return
	}

	// Respond with user
	RespondWithJSON(w, http.StatusOK, user)
}

// LoginUser logs in a user
func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var loginUserParams model.LoginParams

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&loginUserParams); err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to decode request body: %v", err))
		return
	}

	// Login user
	tokens, err := h.userService.LoginUser(r.Context(), loginUserParams)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to login user: %v", err))
		return
	}

	// Respond with tokens
	RespondWithJSON(w, http.StatusOK, tokens)
}

// RequestPasswordReset requests a password reset
func (h *UserHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Email string `json:"email"`
	}

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to decode request body: %v", err))
		return
	}

	// Send reset password email
	if err := h.userService.SendPasswordResetEmail(r.Context(), params.Email); err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to send reset password email: %v", err))
		return
	}

	// Respond with success message
	RespondWithSuccess(w, http.StatusOK, "Reset password email sent successfully")
}

// ResetPassword resets a user's password
func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		token := r.URL.Query().Get("token")
		if token == "" {
			RespondWithError(w, http.StatusBadRequest, "Token is required")
			return
		}

		w.Header().Set("Content-Type", "text/html")
		tmpl := template.Must(template.ParseFiles("pkg/templates/reset-password.html"))
		data := map[string]interface{}{
			"Token": token,
		}
		if err := tmpl.Execute(w, data); err != nil {
			RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to open reset password page: %v", err))
			return
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid form data")
			return
		}

		token := r.FormValue("token")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		if token == "" {
			RespondWithError(w, http.StatusBadRequest, "Token is required")
			return
		}
		if password != confirmPassword {
			RespondWithError(w, http.StatusBadRequest, "Passwords do not match")
			return
		}

		if err := h.userService.UpdateUserPassword(r.Context(), token, password); err != nil {
			RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to reset password: %v", err))
			return
		}

		RespondWithSuccess(w, http.StatusOK, "Password reset successfully")
	}
}

// RefreshToken refreshes a user's access token
func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.Header.Get("Authorization")
	if refreshToken == "" {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tokens, err := h.userService.RefreshToken(r.Context(), refreshToken)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to refresh token: %v", err))
		return
	}

	RespondWithJSON(w, http.StatusOK, tokens)
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.userService.DeactivateUser(r.Context(), userID); err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete user: %v", err))
		return
	}

	RespondWithSuccess(w, http.StatusOK, "User deleted successfully")
}

// UpdateUserProfile updates a user's profile
func (h *UserHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	var params model.UpdateUserProfileParams

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to decode request body: %v", err))
		return
	}

	user, err := h.userService.UpdateUserProfile(r.Context(), params)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update user profile: %v", err))
		return
	}

	RespondWithJSON(w, http.StatusOK, user)
}

// GetUserProfile gets a user's profile
func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUserProfile(r.Context(), userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get user profile: %v", err))
		return
	}

	RespondWithJSON(w, http.StatusOK, user)
}

// ListUsers lists all users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page, pageSize, err := GetPageAndPageSize(pageStr, pageSizeStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid page or page size: %v", err))
		return
	}

	users, err := h.userService.ListUsers(r.Context(), pageSize, page)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list users: %v", err))
		return
	}

	RespondWithJSON(w, http.StatusOK, users)
}
