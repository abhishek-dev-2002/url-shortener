package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"

	"github.com/abhishekmaurya/url-shortner/repo"
	"github.com/abhishekmaurya/url-shortner/utils"
)

type contextKey string

const (
	requestIDKey   contextKey = "request_id"
	requestBodyKey contextKey = "request_body"
)

type GenericHTTPResponsePayload struct {
	Message any `json:"message,omitempty"`
}

func CommonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				utils.Error("panic recovered", "panic", recovered, "stack", string(debug.Stack()))
				sendResponse(w, r, nil, utils.InternalError("internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(recorder, r)

		utils.Info("request completed",
			"requestId", GetRequestID(r),
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.statusCode,
			"latency_ms", time.Since(start).Milliseconds(),
		)
	})
}

func GenerateHandlerFunc(repoMgr *repo.RepositoryManager, handler func(*http.Request) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = addRequestContextIfNotPresent(r, repoMgr)
		response, err := handler(r)
		sendResponse(w, r, response, err)
	}
}

func DecoderMiddleware(newBody func() any, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := newBody()
		if err := json.NewDecoder(r.Body).Decode(body); err != nil {
			sendResponse(w, r, nil, utils.BadRequest("invalid request body"))
			return
		}

		ctx := context.WithValue(r.Context(), requestBodyKey, body)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func GetRequestBody(r *http.Request) any {
	return r.Context().Value(requestBodyKey)
}

func GetRequestID(r *http.Request) string {
	value, _ := r.Context().Value(requestIDKey).(string)
	return value
}

func GetMessagePayload(message string) GenericHTTPResponsePayload {
	if message == "" {
		message = "Success"
	}
	return GenericHTTPResponsePayload{Message: message}
}

func addRequestContextIfNotPresent(r *http.Request, _ *repo.RepositoryManager) *http.Request {
	if GetRequestID(r) != "" {
		return r
	}

	updated := r.Header.Get("X-Request-ID")
	if updated == "" {
		updated = uuid.NewString()
	}
	return r.WithContext(context.WithValue(r.Context(), requestIDKey, updated))
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(statusCode int) {
	sr.statusCode = statusCode
	sr.ResponseWriter.WriteHeader(statusCode)
}

func sendResponse(w http.ResponseWriter, r *http.Request, response any, err error) {
	if err != nil {
		appErr, ok := err.(*utils.AppError)
		if !ok {
			appErr = utils.InternalError(fmt.Sprintf("unexpected error: %v", err))
		}
		utils.WriteError(w, appErr, GetRequestID(r))
		return
	}

	if redirect, ok := response.(*utils.RedirectResponse); ok {
		http.Redirect(w, redirect.Request, redirect.URL, redirect.StatusCode)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, response, "Success", GetRequestID(r))
}
