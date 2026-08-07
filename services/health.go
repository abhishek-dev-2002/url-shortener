package services

import (
"context"
"database/sql"
"net/http"
"time"

"github.com/gin-gonic/gin"
)

// DBPinger is the interface that the health handler needs — just a ping.
// The repo.RepositoryManager satisfies this via its DB() method.
type DBPinger interface {
	DB() *sql.DB
}

// HealthHandler provides health check endpoints.
type HealthHandler struct {
	pinger DBPinger
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(pinger DBPinger) *HealthHandler {
	return &HealthHandler{pinger: pinger}
}

// HealthCheck returns the health status of the service and its dependencies.
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	dbStatus := "up"
	if err := h.pinger.DB().PingContext(ctx); err != nil {
		dbStatus = "down"
	}

	status := http.StatusOK
	overall := "healthy"
	if dbStatus == "down" {
		status = http.StatusServiceUnavailable
		overall = "unhealthy"
	}

	c.JSON(status, gin.H{
"status": overall,
"dependencies": gin.H{
"postgres": dbStatus,
},
})
}
