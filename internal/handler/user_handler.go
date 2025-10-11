package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/yogawahyudi7/go-otel/internal/domain"
	"github.com/yogawahyudi7/go-otel/internal/service"
	"github.com/yogawahyudi7/go-otel/pkg/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	service service.UserService
	tracer  trace.Tracer
}

// NewUserHandler creates a new user handler
func NewUserHandler(service service.UserService, tracer trace.Tracer) *UserHandler {
	return &UserHandler{
		service: service,
		tracer:  tracer,
	}
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c echo.Context) error {
	ctx, span := h.tracer.Start(c.Request().Context(), "UserHandler.CreateUser")
	defer span.End()

	var req domain.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		span.RecordError(err)
		logger.Error("Failed to bind request", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	user, err := h.service.CreateUser(ctx, &req)
	if err != nil {
		span.RecordError(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user",
		})
	}

	span.SetAttributes(attribute.String("user.id", user.ID.String()))
	return c.JSON(http.StatusCreated, user)
}

// GetUser handles GET /users/:id
func (h *UserHandler) GetUser(c echo.Context) error {
	ctx, span := h.tracer.Start(c.Request().Context(), "UserHandler.GetUser")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		span.RecordError(err)
		logger.Error("Invalid user ID", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid user ID",
		})
	}

	span.SetAttributes(attribute.String("user.id", id.String()))

	user, err := h.service.GetUser(ctx, id)
	if err != nil {
		span.RecordError(err)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "User not found",
		})
	}

	return c.JSON(http.StatusOK, user)
}

// ListUsers handles GET /users
func (h *UserHandler) ListUsers(c echo.Context) error {
	ctx, span := h.tracer.Start(c.Request().Context(), "UserHandler.ListUsers")
	defer span.End()

	// Parse pagination parameters
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	span.SetAttributes(
		attribute.Int("pagination.limit", limit),
		attribute.Int("pagination.offset", offset),
	)

	users, total, err := h.service.ListUsers(ctx, limit, offset)
	if err != nil {
		span.RecordError(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to list users",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":   users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateUser handles PUT /users/:id
func (h *UserHandler) UpdateUser(c echo.Context) error {
	ctx, span := h.tracer.Start(c.Request().Context(), "UserHandler.UpdateUser")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		span.RecordError(err)
		logger.Error("Invalid user ID", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid user ID",
		})
	}

	span.SetAttributes(attribute.String("user.id", id.String()))

	var req domain.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		span.RecordError(err)
		logger.Error("Failed to bind request", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	user, err := h.service.UpdateUser(ctx, id, &req)
	if err != nil {
		span.RecordError(err)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "User not found",
		})
	}

	return c.JSON(http.StatusOK, user)
}

// DeleteUser handles DELETE /users/:id
func (h *UserHandler) DeleteUser(c echo.Context) error {
	ctx, span := h.tracer.Start(c.Request().Context(), "UserHandler.DeleteUser")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		span.RecordError(err)
		logger.Error("Invalid user ID", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid user ID",
		})
	}

	span.SetAttributes(attribute.String("user.id", id.String()))

	if err := h.service.DeleteUser(ctx, id); err != nil {
		span.RecordError(err)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "User not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}
