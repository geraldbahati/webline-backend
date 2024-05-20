package handlers

import (
	"encoding/json"
	"fmt"
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

func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	// params
	var registerUserParams model.RegisterUserParams

	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&registerUserParams); err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to decode request body: %v", err))
		return
	}

	// create user
	user, err := h.userService.CreateUser(r.Context(), registerUserParams)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create user: %v", err))
		return
	}

	// respond with user
	RespondWithJSON(w, http.StatusOK, user)
}
