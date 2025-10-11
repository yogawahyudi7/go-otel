package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yogawahyudi7/go-otel/internal/domain"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindAll(ctx context.Context, limit, offset int) ([]*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

type userRepository struct {
	db     *gorm.DB
	tracer trace.Tracer
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB, tracer trace.Tracer) UserRepository {
	return &userRepository{
		db:     db,
		tracer: tracer,
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	ctx, span := r.tracer.Start(ctx, "UserRepository.Create")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.email", user.Email),
		attribute.String("user.name", user.Name),
	)

	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create user: %w", err)
	}

	span.SetAttributes(attribute.String("user.id", user.ID.String()))
	return nil
}

// FindByID finds a user by ID
func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.FindByID")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", id.String()))

	var user domain.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			span.SetAttributes(attribute.Bool("user.found", false))
			return nil, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	span.SetAttributes(attribute.Bool("user.found", true))
	return &user, nil
}

// FindAll finds all users with pagination
func (r *userRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.FindAll")
	defer span.End()

	span.SetAttributes(
		attribute.Int("pagination.limit", limit),
		attribute.Int("pagination.offset", offset),
	)

	var users []*domain.User
	if err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to find users: %w", err)
	}

	span.SetAttributes(attribute.Int("users.count", len(users)))
	return users, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	ctx, span := r.tracer.Start(ctx, "UserRepository.Update")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", user.ID.String()))

	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete soft deletes a user
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "UserRepository.Delete")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", id.String()))

	if err := r.db.WithContext(ctx).Delete(&domain.User{}, "id = ?", id).Error; err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// Count returns the total number of users
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.Count")
	defer span.End()

	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.User{}).Count(&count).Error; err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	span.SetAttributes(attribute.Int64("users.total", count))
	return count, nil
}
