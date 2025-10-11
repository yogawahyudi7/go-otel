package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.uber.org/zap"

	"github.com/yogawahyudi7/go-otel/config"
	"github.com/yogawahyudi7/go-otel/internal/domain"
	"github.com/yogawahyudi7/go-otel/internal/handler"
	"github.com/yogawahyudi7/go-otel/internal/middleware"
	"github.com/yogawahyudi7/go-otel/internal/repository"
	"github.com/yogawahyudi7/go-otel/internal/service"
	"github.com/yogawahyudi7/go-otel/pkg/database"
	"github.com/yogawahyudi7/go-otel/pkg/logger"
	"github.com/yogawahyudi7/go-otel/pkg/telemetry"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(cfg.Log.Level, cfg.Log.Format); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("Failed to sync logger: %v\n", err)
		}
	}()

	logger.Info("Starting application",
		zap.String("name", cfg.App.Name),
		zap.String("env", cfg.App.Env),
		zap.String("port", cfg.App.Port),
	)

	// Initialize context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize OpenTelemetry
	if err := telemetry.Init(ctx, &cfg.Otel); err != nil {
		logger.Fatal("Failed to initialize telemetry", zap.Error(err))
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := telemetry.Shutdown(shutdownCtx); err != nil {
			logger.Error("Failed to shutdown telemetry", zap.Error(err))
		}
	}()

	// Connect to database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := database.Close(); err != nil {
			logger.Error("Failed to close database connection", zap.Error(err))
		}
	}()

	// Auto migrate database
	if err := db.AutoMigrate(&domain.User{}); err != nil {
		logger.Fatal("Failed to migrate database", zap.Error(err))
	}
	logger.Info("Database migration completed")

	// Initialize tracer
	tracer := telemetry.GetTracer("go-otel-api")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db, tracer)

	// Initialize services
	userService := service.NewUserService(userRepo, tracer)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService, tracer)
	healthHandler := handler.NewHealthHandler()

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Set custom error handler
	e.HTTPErrorHandler = middleware.ErrorHandler

	// Middleware
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.RequestID())
	e.Use(echomiddleware.CORS())
	e.Use(middleware.RequestLogger())

	// OpenTelemetry middleware
	if cfg.Otel.Enabled {
		e.Use(otelecho.Middleware(cfg.Otel.ServiceName))
	}

	// Health check routes
	e.GET("/health/live", healthHandler.Liveness)
	e.GET("/health/ready", healthHandler.Readiness)
	e.GET("/health/startup", healthHandler.Startup)

	// API routes
	api := e.Group("/api/v1")
	{
		// User routes
		users := api.Group("/users")
		users.POST("", userHandler.CreateUser)
		users.GET("", userHandler.ListUsers)
		users.GET("/:id", userHandler.GetUser)
		users.PUT("/:id", userHandler.UpdateUser)
		users.DELETE("/:id", userHandler.DeleteUser)
	}

	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%s", cfg.App.Port)
		logger.Info("Server starting", zap.String("address", addr))

		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
}
