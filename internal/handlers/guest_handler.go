package handlers

import (
	"encoding/json"
	"net/http"
	"weblineBackend/pkg/utils"

	"go.uber.org/zap"
)

type GuestHandler struct {
	logger *zap.Logger
}

func NewGuestHandler(logger *zap.Logger) *GuestHandler {
	return &GuestHandler{logger: logger}
}

type GuestTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpiresAt    int64  `json:"expiresAt"`
}

func (h *GuestHandler) GenerateGuestTokenHandler(w http.ResponseWriter, r *http.Request) {
	accessToken, refreshToken, refreshTokenExpiry, err := utils.GenerateGuestTokens()
	if err != nil {
		http.Error(w, "Failed to generate guest tokens", http.StatusInternalServerError)
		return
	}

	response := GuestTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    refreshTokenExpiry.Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
