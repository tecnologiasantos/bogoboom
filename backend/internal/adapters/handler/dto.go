// Package handler holds the HTTP layer (Fiber v3). Handlers parse and answer; the rules live in
// internal/core/service.
package handler

// ErrorResponse is the body of every error. Details maps a field to its problem on validation errors.
type ErrorResponse struct {
	Error   bool              `json:"error"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}
