package utils

import "github.com/gin-gonic/gin"

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

// SendSuccess sends a standardized success JSON response.
func SendSuccess(c *gin.Context, status int, payload any, message string) {
	c.JSON(status, SuccessResponse{
		RequestID: c.GetString("requestId"),
		Payload:   payload,
		Message:   message,
	})
}

// SendError sends a standardized error JSON response.
func SendError(c *gin.Context, err *AppError) {
	c.JSON(err.HTTPStatus, ErrorResponse{
		RequestID: c.GetString("requestId"),
		Error: ErrorPayload{
			Code:    err.Code,
			Message: err.Message,
		},
	})
}
