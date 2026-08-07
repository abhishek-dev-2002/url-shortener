package utils

import "net/http"

// AppError is the standard application error with code, message, and HTTP status.
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

// Error constructors

func BadRequest(message string) *AppError {
	return &AppError{
		Code:       "BAD_REQUEST",
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

func NotFound(message string) *AppError {
	return &AppError{
		Code:       "NOT_FOUND",
		Message:    message,
		HTTPStatus: http.StatusNotFound,
	}
}

func Conflict(message string) *AppError {
	return &AppError{
		Code:       "CONFLICT",
		Message:    message,
		HTTPStatus: http.StatusConflict,
	}
}

func InternalError(message string) *AppError {
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
	}
}
