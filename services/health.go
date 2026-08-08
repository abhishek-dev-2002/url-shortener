package services

import (
	"context"
	"database/sql"
	"net/http"
	"time"
)

// DBPinger is the minimal contract needed by the health service.
type DBPinger interface {
	DB() *sql.DB
}

type HealthService struct {
	pinger DBPinger
}

// NewHealthService creates a new health service.
func NewHealthService(pinger DBPinger) *HealthService {
	return &HealthService{pinger: pinger}
}

// HealthCheck reports the status of the database dependency.
func (h *HealthService) HealthCheck(r *http.Request) (any, error) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	dbStatus := "up"
	if err := h.pinger.DB().PingContext(ctx); err != nil {
		dbStatus = "down"
	}

	status := "healthy"
	if dbStatus == "down" {
		status = "unhealthy"
	}

	return map[string]any{
		"status": status,
		"dependencies": map[string]any{
			"postgres": dbStatus,
		},
	}, nil
}
