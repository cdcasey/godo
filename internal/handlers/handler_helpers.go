package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *TodoHandler) respondJson(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Error encoding response", "error", http.StatusInternalServerError)
	}
}
