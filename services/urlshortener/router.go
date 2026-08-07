package urlshortener

import (
"github.com/gin-gonic/gin"

"github.com/abhishekmaurya/url-shortner/services"
)

// SetupRouter configures all routes for the application.
func SetupRouter(urlHandler *Handler, healthHandler *services.HealthHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(services.RequestIDMiddleware())
	router.Use(services.RequestLoggerMiddleware())

	// Health check
	router.GET("/health", healthHandler.HealthCheck)

	// API v1
	v1 := router.Group("/api/v1")
	{
		v1.POST("/shorten", urlHandler.Shorten)
	}

	// Redirect at root level
	router.GET("/:code", urlHandler.Redirect)

	return router
}
