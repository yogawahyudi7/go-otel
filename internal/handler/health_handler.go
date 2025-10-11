package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yogawahyudi7/go-otel/pkg/database"
	"github.com/yogawahyudi7/go-otel/pkg/logger"
	"go.uber.org/zap"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string            `json:"status"`
	Version string            `json:"version"`
	Checks  map[string]string `json:"checks"`
	Uptime  string            `json:"uptime"`
	Time    string            `json:"time"`
}

var startTime = time.Now()

// Liveness handles GET /health/live
// This endpoint is used by Kubernetes liveness probe
func (h *HealthHandler) Liveness(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "alive",
	})
}

// Readiness handles GET /health/ready
// This endpoint is used by Kubernetes readiness probe
func (h *HealthHandler) Readiness(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]string)
	allHealthy := true

	// Check database connection
	if err := database.HealthCheck(ctx); err != nil {
		checks["database"] = "unhealthy"
		allHealthy = false
		logger.Error("Database health check failed", zap.Error(err))
	} else {
		checks["database"] = "healthy"
	}

	status := "ready"
	statusCode := http.StatusOK

	if !allHealthy {
		status = "not_ready"
		statusCode = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:  status,
		Version: "1.0.0",
		Checks:  checks,
		Uptime:  time.Since(startTime).String(),
		Time:    time.Now().UTC().Format(time.RFC3339),
	}

	return c.JSON(statusCode, response)
}

// Startup handles GET /health/startup
// This endpoint is used by Kubernetes startup probe
func (h *HealthHandler) Startup(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	// Check if database is ready
	if err := database.HealthCheck(ctx); err != nil {
		logger.Error("Startup check failed - database not ready", zap.Error(err))
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "starting",
			"error":  "database not ready",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "started",
	})
}
