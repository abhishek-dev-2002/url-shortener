package services

import (
"time"

"github.com/gin-gonic/gin"
"github.com/google/uuid"

"github.com/abhishekmaurya/url-shortner/utils"
)

// RequestIDMiddleware adds a unique request ID to each request context.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("requestId", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// RequestLoggerMiddleware logs request details.
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		utils.Info("request completed",
"requestId", c.GetString("requestId"),
"method", c.Request.Method,
"path", c.Request.URL.Path,
"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
		)
	}
}
