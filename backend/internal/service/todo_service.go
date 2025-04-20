package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/Sosokker/todolist-backend/internal/domain"
	"github.com/Sosokker/todolist-backend/internal/repository"
	"github.com/google/uuid"
)

type todoService struct {
	todoRepo       repository.TodoRepository
	tagService     TagService
	subtaskService SubtaskService
	storageService FileStorageService
	logger         *slog.Logger
}

// NewTodoService creates a new TodoService
func NewTodoService(
	todoRepo repository.TodoRepository,
	tagService TagService,
	subtaskService SubtaskService,
	storageService FileStorageService,
) TodoService {
	return &todoService{
		todoRepo:       todoRepo,
		tagService:     tagService,
		subtaskService: subtaskService,
		storageService: storageService,
		logger:         slog.Default().With("service", "todo"),
	}
}

func (s *todoService) CreateTodo(ctx context.Context, userID uuid.UUID, input CreateTodoInput) (*domain.Todo, error) {
	// Validate input
	if input.Title == "" {
		return nil, fmt.Errorf("title is required: %w", domain.ErrValidation)
	}

	// Validate associated Tag IDs belong to the user
	if len(input.TagIDs) > 0 {
		if err := s.tagService.ValidateUserTags(ctx, userID, input.TagIDs); err != nil {
			return nil, err // Propagate validation or not found errors
		}
	}

	// Set default status if not provided
	status := domain.StatusPending
	if input.Status != nil {
		status = *input.Status
	}

	newTodo := &domain.Todo{
		UserID:        userID,
		Title:         input.Title,
		Description:   input.Description,
		Status:        status,
		Deadline:      input.Deadline,
		TagIDs:        input.TagIDs,
		AttachmentUrl: nil, // No attachment on creation
	}

	createdTodo, err := s.todoRepo.Create(ctx, newTodo)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create todo in repo", "error", err, "userId", userID)
		return nil, domain.ErrInternalServer
	}

	// Associate Tags if provided (after Todo creation)
	if len(input.TagIDs) > 0 {
		if err = s.todoRepo.SetTags(ctx, createdTodo.ID, input.TagIDs); err != nil {
			s.logger.ErrorContext(ctx, "Failed to associate tags during todo creation", "error", err, "todoId", createdTodo.ID)
			_ = s.todoRepo.Delete(ctx, createdTodo.ID, userID) // Best effort cleanup
			return nil, domain.ErrInternalServer
		}
		createdTodo.TagIDs = input.TagIDs
	}

	return createdTodo, nil
}

func (s *todoService) GetTodoByID(ctx context.Context, todoID, userID uuid.UUID) (*domain.Todo, error) {
	todo, err := s.todoRepo.GetByID(ctx, todoID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.logger.WarnContext(ctx, "Todo not found or forbidden", "todoId", todoID, "userId", userID)
			return nil, domain.ErrNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to get todo from repo", "error", err, "todoId", todoID, "userId", userID)
		return nil, domain.ErrInternalServer
	}

	// Eager load associated Tags and Subtasks
	tags, err := s.todoRepo.GetTags(ctx, todoID)
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to get tags for todo", "error", err, "todoId", todoID)
	} else {
		todo.Tags = tags
		todo.TagIDs = make([]uuid.UUID, 0, len(tags))
		for _, tag := range tags {
			todo.TagIDs = append(todo.TagIDs, tag.ID)
		}
	}

	subtasks, err := s.subtaskService.ListByTodo(ctx, todoID, userID)
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to get subtasks for todo", "error", err, "todoId", todoID)
	} else {
		todo.Subtasks = subtasks
	}

	// Note: todo.Attachments currently holds storage IDs (paths).
	// The handler will call GetAttachmentURLs to convert these to full URLs for the API response.

	return todo, nil
}

func (s *todoService) ListUserTodos(ctx context.Context, userID uuid.UUID, input ListTodosInput) ([]domain.Todo, error) {
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Offset < 0 {
		input.Offset = 0
	}

	repoParams := repository.ListTodosParams{
		UserID:         userID,
		Status:         input.Status,
		TagID:          input.TagID,
		DeadlineBefore: input.DeadlineBefore,
		DeadlineAfter:  input.DeadlineAfter,
		ListParams: repository.ListParams{
			Limit:  input.Limit,
			Offset: input.Offset,
		},
	}

	todos, err := s.todoRepo.ListByUser(ctx, repoParams)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to list todos from repo", "error", err, "userId", userID)
		return nil, domain.ErrInternalServer
	}

	// Optional: Eager load Tags for each Todo in the list efficiently
	// 1. Collect all Todo IDs
	// 2. Make one batch query to get all tags for these todos (e.g., WHERE todo_id IN (...))
	// 3. Map tags back to their respective todos
	// See todo_repo.go for implementation notes.
	// This avoids N+1 queries.

	return todos, nil
}

func (s *todoService) UpdateTodo(ctx context.Context, todoID, userID uuid.UUID, input UpdateTodoInput) (*domain.Todo, error) {
	existingTodo, err := s.todoRepo.GetByID(ctx, todoID, userID)
	if err != nil {
		return nil, err
	}

	updateData := &domain.Todo{
		ID:            existingTodo.ID,
		UserID:        existingTodo.UserID,
		Title:         existingTodo.Title,
		Description:   existingTodo.Description,
		Status:        existingTodo.Status,
		Deadline:      existingTodo.Deadline,
		TagIDs:        existingTodo.TagIDs,
		AttachmentUrl: existingTodo.AttachmentUrl, // Single attachment URL
	}

	updated := false

	if input.Title != nil {
		if err := ValidateTodoTitle(*input.Title); err != nil {
			return nil, err
		}
		updateData.Title = *input.Title
		updated = true
	}
	if input.Description != nil {
		updateData.Description = input.Description
		updated = true
	}
	if input.Status != nil {
		updateData.Status = *input.Status
		updated = true
	}
	if input.Deadline != nil {
		updateData.Deadline = input.Deadline
		updated = true
	}

	tagsUpdated := false
	if input.TagIDs != nil {
		if len(*input.TagIDs) > 0 {
			if err := s.tagService.ValidateUserTags(ctx, userID, *input.TagIDs); err != nil {
				return nil, err
			}
		}
		err = s.todoRepo.SetTags(ctx, todoID, *input.TagIDs)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to update tags for todo", "error", err, "todoId", todoID)
			return nil, domain.ErrInternalServer
		}
		tagsUpdated = true
	}

	// Update the core fields if anything changed
	var updatedRepoTodo *domain.Todo
	if updated {
		updatedRepoTodo, err = s.todoRepo.Update(ctx, todoID, userID, updateData)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to update todo in repo", "error", err, "todoId", todoID)
			return nil, domain.ErrInternalServer
		}
	} else {
		// If only tags were updated, we still need the latest full todo data
		updatedRepoTodo = existingTodo
	}

	// If tags were updated, reload the full todo to get the updated TagIDs array
	if tagsUpdated {
		reloadedTodo, reloadErr := s.GetTodoByID(ctx, todoID, userID)
		if reloadErr != nil {
			s.logger.WarnContext(ctx, "Failed to reload todo after tag update, returning potentially stale data", "error", reloadErr, "todoId", todoID)
			// Return the todo data we have, even if tags might be slightly out of sync temporarily
			if updatedRepoTodo != nil {
				updatedRepoTodo.TagIDs = *input.TagIDs // Manually set IDs based on input
				return updatedRepoTodo, nil
			}
			return existingTodo, nil // Fallback
		}
		return reloadedTodo, nil
	}

	return updatedRepoTodo, nil // Return the result from repo Update or existing if only tags changed
}

func (s *todoService) DeleteTodo(ctx context.Context, todoID, userID uuid.UUID) error {
	existingTodo, err := s.todoRepo.GetByID(ctx, todoID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil // Already deleted or doesn't exist/belong to user
		}
		return err // Internal error
	}

	// Delete the Todo record from the database first
	err = s.todoRepo.Delete(ctx, todoID, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to delete todo from repo", "error", err, "todoId", todoID, "userId", userID)
		return domain.ErrInternalServer
	}

	// If there is an attachment, attempt to delete it from storage (best effort)
	if existingTodo.AttachmentUrl != nil {
		storageID := *existingTodo.AttachmentUrl
		deleteCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		if err := s.storageService.Delete(deleteCtx, storageID); err != nil {
			s.logger.WarnContext(ctx, "Failed to delete attachment file during todo deletion", "error", err, "storageId", storageID, "todoId", todoID)
		} else {
			s.logger.InfoContext(ctx, "Deleted attachment file during todo deletion", "storageId", storageID, "todoId", todoID)
		}
	}

	s.logger.InfoContext(ctx, "Successfully deleted todo and attempted attachment cleanup", "todoId", todoID, "userId", userID)
	return nil
}

// --- Subtask Delegation Methods ---

func (s *todoService) ListSubtasks(ctx context.Context, todoID, userID uuid.UUID) ([]domain.Subtask, error) {
	_, err := s.todoRepo.GetByID(ctx, todoID, userID)
	if err != nil {
		return nil, err
	}
	return s.subtaskService.ListByTodo(ctx, todoID, userID)
}

func (s *todoService) CreateSubtask(ctx context.Context, todoID, userID uuid.UUID, input CreateSubtaskInput) (*domain.Subtask, error) {
	_, err := s.todoRepo.GetByID(ctx, todoID, userID)
	if err != nil {
		return nil, err
	}
	return s.subtaskService.Create(ctx, todoID, input)
}

func (s *todoService) UpdateSubtask(ctx context.Context, todoID, subtaskID, userID uuid.UUID, input UpdateSubtaskInput) (*domain.Subtask, error) {
	// Check if parent todo belongs to user first (optional but safer)
	_, err := s.todoRepo.GetByID(ctx, todoID, userID)
	if err != nil {
		return nil, err
	}
	// Subtask service's GetByID/Update methods inherently check ownership via JOINs
	return s.subtaskService.Update(ctx, subtaskID, userID, input)
}

func (s *todoService) DeleteSubtask(ctx context.Context, todoID, subtaskID, userID uuid.UUID) error {
	// Check if parent todo belongs to user first (optional but safer)
	_, err := s.todoRepo.GetByID(ctx, todoID, userID)
	if err != nil {
		return err
	}
	// Subtask service's Delete method inherently checks ownership via JOINs
	return s.subtaskService.Delete(ctx, subtaskID, userID)
}

// --- Attachment Methods (Simplified) ---

func (s *todoService) AddAttachment(ctx context.Context, todoID, userID uuid.UUID, fileName string, fileSize int64, fileContent io.Reader) (*domain.Todo, error) {
	_, err := s.todoRepo.GetByID(ctx, todoID, userID)
	if err != nil {
		return nil, err
	}

	storageID, _, err := s.storageService.Upload(ctx, userID, todoID, fileName, fileContent, fileSize)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to upload attachment", "error", err, "todoId", todoID)
		return nil, err
	}

	// Construct the public URL for the uploaded file in GCS
	publicURL, err := s.storageService.GetURL(ctx, storageID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate public URL for attachment", "error", err, "todoId", todoID, "storageId", storageID)
		return nil, err
	}

	if err := s.todoRepo.UpdateAttachmentURL(ctx, todoID, userID, &publicURL); err != nil {
		s.logger.ErrorContext(ctx, "Failed to update attachment URL in repo", "error", err, "todoId", todoID)
		return nil, err
	}

	s.logger.InfoContext(ctx, "Attachment added successfully", "todoId", todoID, "storageId", storageID)

	return s.GetTodoByID(ctx, todoID, userID)
}

func (s *todoService) DeleteAttachment(ctx context.Context, todoID, userID uuid.UUID) error {
	_, err := s.todoRepo.GetByID(ctx, todoID, userID)
	if err != nil {
		return err
	}

	if err := s.todoRepo.UpdateAttachmentURL(ctx, todoID, userID, nil); err != nil {
		s.logger.ErrorContext(ctx, "Failed to update attachment URL in repo", "error", err, "todoId", todoID)
		return err
	}

	s.logger.InfoContext(ctx, "Attachment deleted successfully", "todoId", todoID)
	return nil
}
