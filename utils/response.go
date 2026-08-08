package utils

import (
	"encoding/json"
	"net/http"
)

// SuccessResponse is the standard success response format.
type SuccessResponse struct {
	RequestID string `json:"requestId"`
	Payload   any    `json:"payload"`
	Message   string `json:"message"`
}

// ErrorResponse is the standard error response format.
type ErrorResponse struct {
	RequestID string       `json:"requestId"`
	Error     ErrorPayload `json:"error"`
}

// ErrorPayload holds the error details.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type RedirectResponse struct {
	Request    *http.Request
	URL        string
	StatusCode int
}

// WriteSuccess sends a standardized success JSON response.
func WriteSuccess(w http.ResponseWriter, status int, payload any, message, requestID string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(SuccessResponse{
		RequestID: requestID,
		Payload:   payload,
		Message:   message,
	})
}

// WriteError sends a standardized error JSON response.
func WriteError(w http.ResponseWriter, err *AppError, requestID string) {
	w.WriteHeader(err.HTTPStatus)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		RequestID: requestID,
		Error: ErrorPayload{
			Code:    err.Code,
			Message: err.Message,
		},
	})
}
