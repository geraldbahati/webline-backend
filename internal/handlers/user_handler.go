package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"weblineBackend/internal/app_errors"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/model"
	"weblineBackend/internal/services"
)

type UserHandler struct {
	userService         *services.UserService
	adminRequestService *services.AdminRequestService
	config              *appconfig.Config
}

func NewUserHandler(userService *services.UserService, adminRequestService *services.AdminRequestService, config *appconfig.Config) *UserHandler {
	return &UserHandler{
		userService:         userService,
		adminRequestService: adminRequestService,
		config:              config,
	}
}

// RegisterUser registers a new user
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var registerUserParams model.RegisterUserParams

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&registerUserParams); err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to decode request body: %v", err))
		return
	}

	// Create user
	err := h.userService.CreateUser(r.Context(), registerUserParams)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create user: %v", err))
		return
	}

	// Respond with user
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "User created successfully",
	})
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
		var appErr *app_errors.AppError
		ok := errors.As(err, &appErr)
		if ok {
			RespondWithError(w, http.StatusBadRequest, appErr.Message)
			return
		}
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

	if params.Email == "" {
		RespondWithError(w, http.StatusBadRequest, "Email is required")
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
	var params struct {
		Password string `json:"password"`
		Token    string `json:"token"`
	}
	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to decode request body: %v", err))
		return
	}

	// check if passwords are empty
	if params.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "Password is required")
		return
	}

	// Reset password
	if err := h.userService.UpdateUserPassword(r.Context(), params.Token, params.Password); err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to reset password: %v", err))
		return
	}

	// Respond with success message
	RespondWithSuccess(w, http.StatusOK, "Password reset successfully")
}

// RefreshToken refreshes a user's access token
func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	const bearerSchema = "Bearer "

	refreshToken := r.Header.Get("Authorization")
	if refreshToken == "" {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if len(refreshToken) < len(bearerSchema) || refreshToken[:len(bearerSchema)] != bearerSchema {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	refreshToken = refreshToken[len(bearerSchema):]

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
//func (h *UserHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
//	var params model.UpdateUserProfileParams
//
//	// Decode request body
//	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
//		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to decode request body: %v", err))
//		return
//	}
//
//	user, err := h.userService.UpdateUserProfile(r.Context(), params)
//	if err != nil {
//		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update user profile: %v", err))
//		return
//	}
//
//	RespondWithJSON(w, http.StatusOK, user)
//}

// GetUserProfile gets a user's profile
func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := mux.Vars(r)["id"]
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

// LoginWithGoogle logs in a user with Google
func (h *UserHandler) LoginWithGoogle(w http.ResponseWriter, r *http.Request) {
	var params model.GoogleUser

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to decode request body: %v", err))
		return
	}

	tokens, err := h.userService.LoginWithGoogle(r.Context(), params)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to login with Google: %v", err))
		return
	}

	RespondWithJSON(w, http.StatusOK, tokens)
}

// EmailVerified checks if a user's email is verified
func (h *UserHandler) EmailVerified(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Email string `json:"email"`
	}

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to decode request body: %v", err))
		return
	}

	err := h.userService.EmailVerified(r.Context(), params.Email)
	if err != nil {
		var appErr *app_errors.AppError
		ok := errors.As(err, &appErr)
		if ok {
			if appErr.Code == app_errors.EmailNotVerifiedCode {
				RespondWithJSON(w, http.StatusOK, map[string]interface{}{
					"message": "Email is not verified",
				})
				return
			}
			RespondWithError(w, http.StatusBadRequest, appErr.Message)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to verify email")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Email is verified",
	})
}

// VerifyEmail verifies a user's email
func (h *UserHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		RespondWithError(w, http.StatusBadRequest, "Token is required")
		return
	}

	err := h.userService.VerifyEmail(r.Context(), token)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to verify email: %v", err))
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/email-verified", h.config.FrontendURL), http.StatusSeeOther)
}

// RequestAdminRole requests an admin role
func (h *UserHandler) RequestAdminRole(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Reason string `json:"reason"`
	}

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to decode request body: %v", err))
		return
	}

	userID, ok := r.Context().Value("userId").(uuid.UUID)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	log.Println(userID)

	err := h.adminRequestService.RequestAdminRole(r.Context(), userID, params.Reason)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to request admin role: %v", err))
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Admin role requested successfully")
}

// ApproveAdminRole approves an admin role
func (h *UserHandler) ApproveAdminRole(w http.ResponseWriter, r *http.Request) {

	var params struct {
		Token string `json:"token"`
	}

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to decode request body: %v", err))
		return
	}

	err := h.adminRequestService.ApproveAdminRequest(r.Context(), params.Token)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to approve admin role: %v", err))
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Admin role approved successfully")
}
