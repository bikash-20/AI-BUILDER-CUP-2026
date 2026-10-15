package httpapi

import (
	"encoding/json"
	"net/http"
)

// Error is the standard error envelope returned by all FinGuard APIs.
type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// ErrorEnvelope wraps an Error under the "error" key.
type ErrorEnvelope struct {
	Error Error `json:"error"`
}

// WriteError writes a JSON error envelope with the given status.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorEnvelope{Error: Error{Code: code, Message: message}})
}

// WriteJSON encodes v as JSON with status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
