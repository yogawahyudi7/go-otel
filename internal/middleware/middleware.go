package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/yogawahyudi7/go-otel/pkg/logger"
)

// RequestLogger logs HTTP requests with structured logging
func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process request
			err := next(c)

			// Log request
			req := c.Request()
			res := c.Response()

			fields := []zap.Field{
				zap.String("method", req.Method),
				zap.String("uri", req.RequestURI),
				zap.String("path", c.Path()),
				zap.Int("status", res.Status),
				zap.Int64("size", res.Size),
				zap.Duration("latency", time.Since(start)),
				zap.String("remote_ip", c.RealIP()),
				zap.String("user_agent", req.UserAgent()),
			}

			if err != nil {
				fields = append(fields, zap.Error(err))
				logger.Error("Request failed", fields...)
			} else {
				logger.Info("Request processed", fields...)
			}

			return err
		}
	}
}

// ErrorHandler is a custom error handler for Echo
func ErrorHandler(err error, c echo.Context) {
	code := 500
	message := "Internal server error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message.(string)
	}

	logger.Error("HTTP error",
		zap.Int("status", code),
		zap.String("message", message),
		zap.String("path", c.Path()),
		zap.Error(err),
	)

	if !c.Response().Committed {
		if c.Request().Method == echo.HEAD {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, map[string]interface{}{
				"error":  message,
				"status": code,
				"path":   c.Path(),
			})
		}
		if err != nil {
			logger.Error("Failed to send error response", zap.Error(err))
		}
	}
}
