package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/Sosokker/todolist-backend/internal/domain"
	"github.com/Sosokker/todolist-backend/internal/repository"
	"github.com/google/uuid"
)

type todoService struct {
	todoRepo       repository.TodoRepository
	tagService     TagService     // Depend on TagService for validation
	subtaskService SubtaskService // Depend on SubtaskService
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
		UserID:      userID,
		Title:       input.Title,
		Description: input.Description,
		Status:      status,
		Deadline:    input.Deadline,
		TagIDs:      input.TagIDs,
		Attachments: []string{},
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
		ID:          existingTodo.ID,
		UserID:      existingTodo.UserID,
		Title:       existingTodo.Title,
		Description: existingTodo.Description,
		Status:      existingTodo.Status,
		Deadline:    existingTodo.Deadline,
		Attachments: existingTodo.Attachments,
	}

	updated := false

	if input.Title != nil {
		if *input.Title == "" {
			return nil, fmt.Errorf("title cannot be empty: %w", domain.ErrValidation)
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
		updateData.TagIDs = *input.TagIDs
		tagsUpdated = true
	}

	attachmentsUpdated := false
	if input.Attachments != nil {
		err = s.todoRepo.SetAttachments(ctx, todoID, userID, *input.Attachments)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to update attachments list for todo", "error", err, "todoId", todoID)
			return nil, domain.ErrInternalServer
		}
		updateData.Attachments = *input.Attachments
		attachmentsUpdated = true
	}

	var updatedRepoTodo *domain.Todo
	if updated {
		updatedRepoTodo, err = s.todoRepo.Update(ctx, todoID, userID, updateData)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to update todo in repo", "error", err, "todoId", todoID)
			return nil, domain.ErrInternalServer
		}
	} else {
		updatedRepoTodo = updateData
	}

	if !updated && (tagsUpdated || attachmentsUpdated) {
		updatedRepoTodo.Title = existingTodo.Title
		updatedRepoTodo.Description = existingTodo.Description
	}

	finalTodo, err := s.GetTodoByID(ctx, todoID, userID)
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to reload todo after update, returning partial data", "error", err, "todoId", todoID)
		return updatedRepoTodo, nil
	}

	return finalTodo, nil
}

func (s *todoService) DeleteTodo(ctx context.Context, todoID, userID uuid.UUID) error {
	existingTodo, err := s.todoRepo.GetByID(ctx, todoID, userID)
	if err != nil {
		return err
	}

	attachmentIDsToDelete := existingTodo.Attachments

	err = s.todoRepo.Delete(ctx, todoID, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to delete todo from repo", "error", err, "todoId", todoID, "userId", userID)
		return domain.ErrInternalServer
	}

	for _, storageID := range attachmentIDsToDelete {
		if err := s.storageService.Delete(ctx, storageID); err != nil {
			s.logger.WarnContext(ctx, "Failed to delete attachment file during todo deletion", "error", err, "storageId", storageID, "todoId", todoID)
		}
	}
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
	return s.subtaskService.Update(ctx, subtaskID, userID, input)
}

func (s *todoService) DeleteSubtask(ctx context.Context, todoID, subtaskID, userID uuid.UUID) error {
	return s.subtaskService.Delete(ctx, subtaskID, userID)
}

// --- Attachment Methods --- (Implementation depends on FileStorageService)

func (s *todoService) AddAttachment(ctx context.Context, todoID, userID uuid.UUID, originalFilename string, fileSize int64, fileContent io.Reader) (*domain.AttachmentInfo, error) {
	_, err := s.todoRepo.GetByID(ctx, todoID, userID)
	if err != nil {
		return nil, err
	}

	storageID, contentType, err := s.storageService.Upload(ctx, userID, todoID, originalFilename, fileContent, fileSize)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to upload attachment to storage", "error", err, "todoId", todoID, "fileName", originalFilename)
		return nil, domain.ErrInternalServer
	}

	if err = s.todoRepo.AddAttachment(ctx, todoID, userID, storageID); err != nil {
		s.logger.ErrorContext(ctx, "Failed to add attachment storage ID to todo", "error", err, "todoId", todoID, "storageId", storageID)
		if delErr := s.storageService.Delete(context.Background(), storageID); delErr != nil {
			s.logger.ErrorContext(ctx, "Failed to delete orphaned attachment file after DB error", "deleteError", delErr, "storageId", storageID)
		}
		return nil, domain.ErrInternalServer
	}

	fileURL, _ := s.storageService.GetURL(ctx, storageID)

	return &domain.AttachmentInfo{
		FileID:      storageID,
		FileName:    originalFilename,
		FileURL:     fileURL,
		ContentType: contentType,
		Size:        fileSize,
	}, nil
}

func (s *todoService) DeleteAttachment(ctx context.Context, todoID, userID uuid.UUID, storageID string) error {
	todo, err := s.todoRepo.GetByID(ctx, todoID, userID)
	if err != nil {
		return err
	}

	found := false
	for _, att := range todo.Attachments {
		if att == storageID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("attachment '%s' not found on todo %s: %w", storageID, todoID, domain.ErrNotFound)
	}

	if err = s.todoRepo.RemoveAttachment(ctx, todoID, userID, storageID); err != nil {
		s.logger.ErrorContext(ctx, "Failed to remove attachment ID from todo", "error", err, "todoId", todoID, "storageId", storageID)
		return domain.ErrInternalServer
	}

	if err = s.storageService.Delete(ctx, storageID); err != nil {
		s.logger.ErrorContext(ctx, "Failed to delete attachment file from storage after removing DB ref", "error", err, "storageId", storageID)
		return nil
	}

	return nil
}
