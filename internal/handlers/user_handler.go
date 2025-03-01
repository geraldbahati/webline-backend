// internal/handlers/user.go

package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"weblineBackend/internal/app_errors"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/middleware"
	"weblineBackend/internal/model"
	"weblineBackend/internal/services"

	"github.com/gorilla/mux"
)

// UserHandler handles user-related endpoints
type UserHandler struct {
	userService         *services.UserService
	adminRequestService *services.AdminRequestService
	config              *appconfig.Config
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(userService *services.UserService, adminRequestService *services.AdminRequestService, config *appconfig.Config) *UserHandler {
	return &UserHandler{
		userService:         userService,
		adminRequestService: adminRequestService,
		config:              config,
	}
}

// RegisterUser registers a new user (accessible to unauthenticated users)
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var registerUserParams model.RegisterUserParams

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&registerUserParams); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Create user
	err := h.userService.CreateUser(r.Context(), registerUserParams)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Respond with success message
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "User created successfully",
	})
}

// LoginUser logs in a user (accessible to unauthenticated users)
func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var loginUserParams model.LoginParams

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&loginUserParams); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
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
		RespondWithError(w, http.StatusInternalServerError, "Failed to login user")
		return
	}

	// Respond with tokens
	RespondWithJSON(w, http.StatusOK, tokens)
}

// RequestPasswordReset requests a password reset (accessible to unauthenticated users)
func (h *UserHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Email string `json:"email"`
	}

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if params.Email == "" {
		RespondWithError(w, http.StatusBadRequest, "Email is required")
		return
	}

	// Send reset password email
	if err := h.userService.SendPasswordResetEmail(r.Context(), params.Email); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to send reset password email")
		return
	}

	// Respond with success message
	RespondWithSuccess(w, http.StatusOK, "Reset password email sent successfully")
}

// ResetPassword resets a user's password (accessible to unauthenticated users with valid token)
func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Password string `json:"password"`
		Token    string `json:"token"`
	}
	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Check if password is empty
	if params.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "Password is required")
		return
	}

	// Reset password
	if err := h.userService.UpdateUserPassword(r.Context(), params.Token, params.Password); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to reset password")
		return
	}

	// Respond with success message
	RespondWithSuccess(w, http.StatusOK, "Password reset successfully")
}

// RefreshToken refreshes a user's access token (accessible to authenticated users)
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
		RespondWithError(w, http.StatusUnauthorized, "Failed to refresh token")
		return
	}

	RespondWithJSON(w, http.StatusOK, tokens)
}

// DeleteUser deletes a user (accessible to authenticated users)
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok || user.IsGuest {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID := user.UserID.String()
	if userID == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.userService.DeactivateUser(r.Context(), userID); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "User deleted successfully")
}

// GetUserProfile gets a user's profile (accessible to authenticated users)
func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// get the user id from the url
	userID := mux.Vars(r)["id"]
	if userID == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	profile, err := h.userService.GetUserProfile(r.Context(), userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get user profile")
		return
	}

	RespondWithJSON(w, http.StatusOK, profile)
}

// ListUsers lists all users (accessible to admin users)
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok || user.IsGuest {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page, pageSize, err := GetPageAndPageSize(pageStr, pageSizeStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid page or page size")
		return
	}

	users, err := h.userService.ListUsers(r.Context(), pageSize, page)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	RespondWithJSON(w, http.StatusOK, users)
}

// LoginWithGoogle logs in a user with Google (accessible to unauthenticated users)
func (h *UserHandler) LoginWithGoogle(w http.ResponseWriter, r *http.Request) {
	var params model.GoogleUser

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	tokens, err := h.userService.LoginWithGoogle(r.Context(), params)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to login with Google")
		return
	}

	RespondWithJSON(w, http.StatusOK, tokens)
}

// EmailVerified checks if a user's email is verified (accessible to authenticated users)
func (h *UserHandler) EmailVerified(w http.ResponseWriter, r *http.Request) {
	params := struct {
		Email string `json:"email"`
	}{}

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if params.Email == "" {
		RespondWithError(w, http.StatusBadRequest, "Email is required")
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

	RespondWithSuccess(w, http.StatusOK, "Email is verified")
}

// VerifyEmail verifies a user's email (accessible via email verification link)
func (h *UserHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		RespondWithError(w, http.StatusBadRequest, "Token is required")
		return
	}

	err := h.userService.VerifyEmail(r.Context(), token)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to verify email")
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/email-verified", h.config.FrontendURL), http.StatusSeeOther)
}

// RequestAdminRole requests an admin role (accessible to authenticated users)
func (h *UserHandler) RequestAdminRole(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Reason string `json:"reason"`
	}

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, ok := middleware.GetUser(r.Context())
	if !ok || user.IsGuest {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := h.adminRequestService.RequestAdminRole(r.Context(), user.UserID, params.Reason)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to request admin role")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Admin role requested successfully")
}

// ApproveAdminRole approves an admin role (accessible to admin users)
func (h *UserHandler) ApproveAdminRole(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Token string `json:"token"`
	}

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := h.adminRequestService.ApproveAdminRequest(r.Context(), params.Token)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to approve admin role")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "Admin role approved successfully")
}

// GetUserInfo gets a user's info (accessible to authenticated users)
func (h *UserHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok || user.IsGuest {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID := user.UserID.String()
	if userID == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	userInfo, err := h.userService.GetUserInfo(r.Context(), userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get user info")
		return
	}

	RespondWithJSON(w, http.StatusOK, userInfo)
}

// UpdateUserProfile updates a user's profile (accessible to authenticated users)
func (h *UserHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok || user.IsGuest {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse the multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	params := model.UpdateUserInfoParams{
		FirstName:   r.FormValue("firstName"),
		LastName:    r.FormValue("lastName"),
		PhoneNumber: r.FormValue("phoneNumber"),
		DateOfBirth: r.FormValue("dateOfBirth"),
	}

	file, header, err := r.FormFile("profileImage")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			RespondWithError(w, http.StatusBadRequest, "Profile image is required")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve profile image")
		return
	}
	defer file.Close()

	image := &model.ImageFile{
		File:       file,
		FileHeader: header,
	}

	err = h.userService.UpdateUserInfo(r.Context(), params, image)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update user profile")
		return
	}

	RespondWithSuccess(w, http.StatusOK, "User profile updated successfully")
}
