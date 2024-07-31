package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"weblineBackend/internal/services"
)

type RoleHandler struct {
	roleService *services.RoleService
}

// NewRoleHandler initializes a new RoleHandler with dependency injection for the RoleService
func NewRoleHandler(roleService *services.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// CreateRole creates a new role
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	// Decode the request body into the params struct
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to decode request body: %v", err))
		return
	}

	role, err := h.roleService.CreateRole(r.Context(), params.Name, params.Description)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create role: %v", err))
		return
	}

	RespondWithSuccess(w, http.StatusCreated, fmt.Sprintf("Role %s created", role.Name))
}
