package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yogawahyudi7/go-otel/internal/domain"
	"github.com/yogawahyudi7/go-otel/internal/repository"
	"github.com/yogawahyudi7/go-otel/pkg/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// UserService defines the interface for user business logic
type UserService interface {
	CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.UserResponse, error)
	GetUser(ctx context.Context, id uuid.UUID) (*domain.UserResponse, error)
	ListUsers(ctx context.Context, limit, offset int) ([]*domain.UserResponse, int64, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *domain.UpdateUserRequest) (*domain.UserResponse, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type userService struct {
	repo   repository.UserRepository
	tracer trace.Tracer
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository, tracer trace.Tracer) UserService {
	return &userService{
		repo:   repo,
		tracer: tracer,
	}
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.UserResponse, error) {
	ctx, span := s.tracer.Start(ctx, "UserService.CreateUser")
	defer span.End()

	logger.Info("Creating user", zap.String("email", req.Email))

	user := &domain.User{
		Name:  req.Name,
		Email: req.Email,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		span.RecordError(err)
		logger.Error("Failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logger.Info("User created successfully", zap.String("id", user.ID.String()))
	span.SetAttributes(attribute.String("user.id", user.ID.String()))

	return user.ToResponse(), nil
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(ctx context.Context, id uuid.UUID) (*domain.UserResponse, error) {
	ctx, span := s.tracer.Start(ctx, "UserService.GetUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", id.String()))
	logger.Debug("Getting user", zap.String("id", id.String()))

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		logger.Error("Failed to get user", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		logger.Warn("User not found", zap.String("id", id.String()))
		return nil, fmt.Errorf("user not found")
	}

	return user.ToResponse(), nil
}

// ListUsers retrieves all users with pagination
func (s *userService) ListUsers(ctx context.Context, limit, offset int) ([]*domain.UserResponse, int64, error) {
	ctx, span := s.tracer.Start(ctx, "UserService.ListUsers")
	defer span.End()

	span.SetAttributes(
		attribute.Int("pagination.limit", limit),
		attribute.Int("pagination.offset", offset),
	)

	logger.Debug("Listing users", zap.Int("limit", limit), zap.Int("offset", offset))

	users, err := s.repo.FindAll(ctx, limit, offset)
	if err != nil {
		span.RecordError(err)
		logger.Error("Failed to list users", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		span.RecordError(err)
		logger.Error("Failed to count users", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	responses := make([]*domain.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	span.SetAttributes(
		attribute.Int("users.returned", len(responses)),
		attribute.Int64("users.total", total),
	)

	return responses, total, nil
}

// UpdateUser updates a user
func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, req *domain.UpdateUserRequest) (*domain.UserResponse, error) {
	ctx, span := s.tracer.Start(ctx, "UserService.UpdateUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", id.String()))
	logger.Info("Updating user", zap.String("id", id.String()))

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		logger.Error("Failed to find user", zap.Error(err))
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		logger.Warn("User not found", zap.String("id", id.String()))
		return nil, fmt.Errorf("user not found")
	}

	// Update fields
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := s.repo.Update(ctx, user); err != nil {
		span.RecordError(err)
		logger.Error("Failed to update user", zap.Error(err))
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	logger.Info("User updated successfully", zap.String("id", user.ID.String()))
	return user.ToResponse(), nil
}

// DeleteUser deletes a user
func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracer.Start(ctx, "UserService.DeleteUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", id.String()))
	logger.Info("Deleting user", zap.String("id", id.String()))

	// Check if user exists
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		logger.Error("Failed to find user", zap.Error(err))
		return fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		logger.Warn("User not found", zap.String("id", id.String()))
		return fmt.Errorf("user not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		span.RecordError(err)
		logger.Error("Failed to delete user", zap.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	logger.Info("User deleted successfully", zap.String("id", id.String()))
	return nil
}
