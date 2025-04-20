package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Sosokker/todolist-backend/internal/domain"     // Adjust path
	"github.com/Sosokker/todolist-backend/internal/repository" // Adjust path
	"github.com/google/uuid"
)

type subtaskService struct {
	subtaskRepo repository.SubtaskRepository
	logger      *slog.Logger
}

// NewSubtaskService creates a new SubtaskService
func NewSubtaskService(repo repository.SubtaskRepository /*, todoRepo repository.TodoRepository */) SubtaskService {
	return &subtaskService{
		subtaskRepo: repo,
		logger:      slog.Default().With("service", "subtask"),
	}
}

func (s *subtaskService) Create(ctx context.Context, todoID uuid.UUID, input CreateSubtaskInput) (*domain.Subtask, error) {
	if err := ValidateCreateSubtaskInput(input); err != nil {
		return nil, err
	}
	// Ownership check of parent todo (todoID) should be done *before* calling this method,
	// typically in the TodoService which orchestrates subtask operations.
	// Alternatively, the repository methods should enforce this via joins (as done in the example repo).

	subtask := &domain.Subtask{
		TodoID:      todoID,
		Description: input.Description,
		Completed:   false, // Default on create
	}

	createdSubtask, err := s.subtaskRepo.Create(ctx, subtask)
	if err != nil {
		// Repo handles foreign key violation check returning ErrBadRequest
		if errors.Is(err, domain.ErrBadRequest) {
			s.logger.WarnContext(ctx, "Subtask creation failed, invalid parent todo", "todoId", todoID)
			return nil, err
		}
		s.logger.ErrorContext(ctx, "Failed to create subtask in repo", "error", err, "todoId", todoID)
		return nil, domain.ErrInternalServer
	}

	s.logger.InfoContext(ctx, "Subtask created successfully", "subtaskId", createdSubtask.ID, "todoId", todoID)
	return createdSubtask, nil
}

func (s *subtaskService) GetByID(ctx context.Context, subtaskID, userID uuid.UUID) (*domain.Subtask, error) {
	subtask, err := s.subtaskRepo.GetByID(ctx, subtaskID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.logger.WarnContext(ctx, "Subtask not found or access denied", "subtaskId", subtaskID, "userId", userID)
		} else {
			s.logger.ErrorContext(ctx, "Failed to get subtask by ID from repo", "error", err, "subtaskId", subtaskID, "userId", userID)
			err = domain.ErrInternalServer
		}
		return nil, err
	}
	return subtask, nil
}

func (s *subtaskService) ListByTodo(ctx context.Context, todoID, userID uuid.UUID) ([]domain.Subtask, error) {
	subtasks, err := s.subtaskRepo.ListByTodo(ctx, todoID, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to list subtasks by todo from repo", "error", err, "todoId", todoID, "userId", userID)
		return nil, domain.ErrInternalServer
	}
	s.logger.DebugContext(ctx, "Listed subtasks for todo", "todoId", todoID, "userId", userID, "count", len(subtasks))
	return subtasks, nil
}

func (s *subtaskService) Update(ctx context.Context, subtaskID, userID uuid.UUID, input UpdateSubtaskInput) (*domain.Subtask, error) {
	if err := ValidateUpdateSubtaskInput(input); err != nil {
		return nil, err
	}
	// Get existing first to ensure NotFound/Forbidden is returned correctly before attempting update,
	// and to have the existing data if only partial fields are provided in input.
	existingSubtask, err := s.GetByID(ctx, subtaskID, userID)
	if err != nil {
		return nil, err // Handles NotFound/Forbidden/Internal
	}

	updateData := &domain.Subtask{
		Description: existingSubtask.Description,
		Completed:   existingSubtask.Completed,
	}
	needsUpdate := false

	if input.Description != nil {
		updateData.Description = *input.Description
		needsUpdate = true
	}
	if input.Completed != nil {
		updateData.Completed = *input.Completed
		needsUpdate = true
	}

	if !needsUpdate {
		s.logger.InfoContext(ctx, "No fields provided for subtask update", "subtaskId", subtaskID, "userId", userID)
		return existingSubtask, nil
	}

	updatedSubtask, err := s.subtaskRepo.Update(ctx, subtaskID, userID, updateData)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.logger.WarnContext(ctx, "Subtask update failed, not found or access denied", "subtaskId", subtaskID, "userId", userID)
		} else {
			s.logger.ErrorContext(ctx, "Failed to update subtask in repo", "error", err, "subtaskId", subtaskID, "userId", userID)
			err = domain.ErrInternalServer
		}
		return nil, err
	}

	s.logger.InfoContext(ctx, "Subtask updated successfully", "subtaskId", subtaskID, "userId", userID)
	return updatedSubtask, nil
}

func (s *subtaskService) Delete(ctx context.Context, subtaskID, userID uuid.UUID) error {
	// Check existence and ownership first to return proper NotFound/Forbidden.
	_, err := s.GetByID(ctx, subtaskID, userID)
	if err != nil {
		return err // Handles NotFound/Forbidden/Internal
	}

	err = s.subtaskRepo.Delete(ctx, subtaskID, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to delete subtask from repo", "error", err, "subtaskId", subtaskID, "userId", userID)
		return domain.ErrInternalServer
	}

	s.logger.InfoContext(ctx, "Subtask deleted successfully", "subtaskId", subtaskID, "userId", userID)
	return nil
}
