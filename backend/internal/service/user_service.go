package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Sosokker/todolist-backend/internal/domain"     // Adjust path
	"github.com/Sosokker/todolist-backend/internal/repository" // Adjust path
	"github.com/google/uuid"
)

type userService struct {
	userRepo repository.UserRepository
	logger   *slog.Logger
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		userRepo: repo,
		logger:   slog.Default().With("service", "user"),
	}
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.logger.WarnContext(ctx, "User not found by ID", "userId", id)
			return nil, domain.ErrNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to get user by ID from repo", "error", err, "userId", id)
		return nil, domain.ErrInternalServer
	}
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, userID uuid.UUID, input UpdateUserInput) (*domain.User, error) {
	// GetUserByID handles NotFound/Forbidden error
	existingUser, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Prepare update data DTO for the repository
	updateData := &domain.User{
		// Copy non-updatable fields or handle defaults
	}
	needsUpdate := false

	if input.Username != nil {
		if err := ValidateUsername(*input.Username); err != nil {
			return nil, err
		}
		updateData.Username = *input.Username
		needsUpdate = true
	} else {
		updateData.Username = existingUser.Username
	}

	// TODO: Add logic for other updatable fields (e.g., email)
	// Password updates should involve hashing and likely be a separate endpoint/service method.
	// Email updates might require a verification flow.

	if !needsUpdate {
		s.logger.InfoContext(ctx, "No fields provided for user update", "userId", userID)
		return existingUser, nil
	}

	updatedUser, err := s.userRepo.Update(ctx, userID, updateData)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			s.logger.WarnContext(ctx, "User update conflict", "error", err, "userId", userID, "conflictingUsername", updateData.Username)
			return nil, fmt.Errorf("username '%s' is already taken: %w", updateData.Username, domain.ErrConflict)
		}
		s.logger.ErrorContext(ctx, "Failed to update user in repo", "error", err, "userId", userID)
		return nil, domain.ErrInternalServer
	}

	s.logger.InfoContext(ctx, "User updated successfully", "userId", userID)
	return updatedUser, nil
}
