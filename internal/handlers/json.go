package handlers

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

// RespondWithError responds with an error message and status code
func RespondWithError(w http.ResponseWriter, code int, message string) {
	if code >= http.StatusInternalServerError {
		// log error for server-side issues
		zap.L().Error("Server error", zap.Int("code", code), zap.String("message", message))
	}
	RespondWithJSON(w, code, map[string]string{"error": message})
}

// RespondWithJSON responds with a JSON payload and status code
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error("Failed to marshal JSON", zap.Any("payload", payload), zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// RespondWithSuccess responds with a success message and status code
func RespondWithSuccess(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"message": message})
}
