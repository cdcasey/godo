package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func writeJsonResponse(w http.ResponseWriter, statusCode int, data any, logger *slog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Error encoding response", "error", err)
	}
}
