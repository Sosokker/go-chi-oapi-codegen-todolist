package service

import (
	"context"
	"io"
	"time"

	"github.com/Sosokker/todolist-backend/internal/domain"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// --- Auth Service ---
type SignupCredentials struct {
	Username string
	Email    string
	Password string
}

type LoginCredentials struct {
	Email    string
	Password string
}

type AuthService interface {
	Signup(ctx context.Context, creds SignupCredentials) (*domain.User, error)
	Login(ctx context.Context, creds LoginCredentials) (token string, user *domain.User, err error)
	GenerateJWT(user *domain.User) (string, error)
	ValidateJWT(tokenString string) (*domain.User, error)
	GetGoogleAuthConfig() *oauth2.Config
	HandleGoogleCallback(ctx context.Context, code string) (token string, user *domain.User, err error)
}

// --- User Service ---
type UpdateUserInput struct {
	Username *string
}

type UserService interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, input UpdateUserInput) (*domain.User, error)
}

// --- Tag Service ---
type CreateTagInput struct {
	Name  string
	Color *string
	Icon  *string
}

type UpdateTagInput struct {
	Name  *string
	Color *string
	Icon  *string
}

type TagService interface {
	CreateTag(ctx context.Context, userID uuid.UUID, input CreateTagInput) (*domain.Tag, error)
	GetTagByID(ctx context.Context, tagID, userID uuid.UUID) (*domain.Tag, error)
	ListUserTags(ctx context.Context, userID uuid.UUID) ([]domain.Tag, error)
	UpdateTag(ctx context.Context, tagID, userID uuid.UUID, input UpdateTagInput) (*domain.Tag, error)
	DeleteTag(ctx context.Context, tagID, userID uuid.UUID) error
	ValidateUserTags(ctx context.Context, userID uuid.UUID, tagIDs []uuid.UUID) error
}

// --- Todo Service ---
type CreateTodoInput struct {
	Title       string
	Description *string
	Status      *domain.TodoStatus
	Deadline    *time.Time
	TagIDs      []uuid.UUID
}

type UpdateTodoInput struct {
	Title       *string
	Description *string
	Status      *domain.TodoStatus
	Deadline    *time.Time
	TagIDs      *[]uuid.UUID
	// Attachments are managed via separate endpoints
}

type ListTodosInput struct {
	Status         *domain.TodoStatus
	TagID          *uuid.UUID
	DeadlineBefore *time.Time
	DeadlineAfter  *time.Time
	Limit          int
	Offset         int
}

type TodoService interface {
	CreateTodo(ctx context.Context, userID uuid.UUID, input CreateTodoInput) (*domain.Todo, error)
	GetTodoByID(ctx context.Context, todoID, userID uuid.UUID) (*domain.Todo, error) // Fetches attachment URL
	ListUserTodos(ctx context.Context, userID uuid.UUID, input ListTodosInput) ([]domain.Todo, error)
	UpdateTodo(ctx context.Context, todoID, userID uuid.UUID, input UpdateTodoInput) (*domain.Todo, error)
	DeleteTodo(ctx context.Context, todoID, userID uuid.UUID) error
	// Subtask methods
	ListSubtasks(ctx context.Context, todoID, userID uuid.UUID) ([]domain.Subtask, error)
	CreateSubtask(ctx context.Context, todoID, userID uuid.UUID, input CreateSubtaskInput) (*domain.Subtask, error)
	UpdateSubtask(ctx context.Context, todoID, subtaskID, userID uuid.UUID, input UpdateSubtaskInput) (*domain.Subtask, error)
	DeleteSubtask(ctx context.Context, todoID, subtaskID, userID uuid.UUID) error
	// Attachment methods
	AddAttachment(ctx context.Context, todoID, userID uuid.UUID, fileName string, fileSize int64, fileContent io.Reader) (*domain.Todo, error)
	// Uploads, gets URL, updates Todo, returns updated Todo
	DeleteAttachment(ctx context.Context, todoID, userID uuid.UUID) error // Deletes from storage and clears Todo URL
}

// --- Subtask Service ---
type CreateSubtaskInput struct {
	Description string
}

type UpdateSubtaskInput struct {
	Description *string
	Completed   *bool
}

// SubtaskService operates assuming the parent Todo's ownership has already been verified
type SubtaskService interface {
	Create(ctx context.Context, todoID uuid.UUID, input CreateSubtaskInput) (*domain.Subtask, error)
	GetByID(ctx context.Context, subtaskID, userID uuid.UUID) (*domain.Subtask, error)                          // Still need userID for underlying repo call
	ListByTodo(ctx context.Context, todoID, userID uuid.UUID) ([]domain.Subtask, error)                         // Still need userID for underlying repo call
	Update(ctx context.Context, subtaskID, userID uuid.UUID, input UpdateSubtaskInput) (*domain.Subtask, error) // Still need userID
	Delete(ctx context.Context, subtaskID, userID uuid.UUID) error                                              // Still need userID
}

// FileStorageService defines the interface for handling file uploads and deletions.
type FileStorageService interface {
	// Upload saves the content from the reader and returns a unique storage identifier (e.g., path/key) and the content type.
	Upload(ctx context.Context, userID, todoID uuid.UUID, originalFilename string, reader io.Reader, size int64) (storageID string, contentType string, err error)
	// Delete removes the file associated with the given storage identifier.
	Delete(ctx context.Context, storageID string) error
	// GetURL retrieves a publicly accessible URL for the storage ID (e.g., signed URL for GCS).
	GetURL(ctx context.Context, storageID string) (string, error)
	// GenerateUniqueObjectName creates a unique storage path/name for a file.
	GenerateUniqueObjectName(userID, todoID uuid.UUID, originalFilename string) string
}

// ServiceRegistry bundles services
type ServiceRegistry struct {
	Auth    AuthService
	User    UserService
	Tag     TagService
	Todo    TodoService
	Subtask SubtaskService
	Storage FileStorageService
}
