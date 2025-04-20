package repository

import (
	"context"
	"time"

	"github.com/Sosokker/todolist-backend/internal/domain"
	db "github.com/Sosokker/todolist-backend/internal/repository/sqlc/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Common arguments for list methods
type ListParams struct {
	Limit  int
	Offset int
}

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByGoogleID(ctx context.Context, googleID string) (*domain.User, error)
	Update(ctx context.Context, id uuid.UUID, updateData *domain.User) (*domain.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type TagRepository interface {
	Create(ctx context.Context, tag *domain.Tag) (*domain.Tag, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Tag, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) ([]domain.Tag, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Tag, error)
	Update(ctx context.Context, id, userID uuid.UUID, updateData *domain.Tag) (*domain.Tag, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type ListTodosParams struct {
	UserID         uuid.UUID
	Status         *domain.TodoStatus
	TagID          *uuid.UUID
	DeadlineBefore *time.Time
	DeadlineAfter  *time.Time
	ListParams     // Embed pagination
}

type TodoRepository interface {
	Create(ctx context.Context, todo *domain.Todo) (*domain.Todo, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Todo, error)
	ListByUser(ctx context.Context, params ListTodosParams) ([]domain.Todo, error)
	Update(ctx context.Context, id, userID uuid.UUID, updateData *domain.Todo) (*domain.Todo, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	// Tag associations
	AddTag(ctx context.Context, todoID, tagID uuid.UUID) error
	RemoveTag(ctx context.Context, todoID, tagID uuid.UUID) error
	SetTags(ctx context.Context, todoID uuid.UUID, tagIDs []uuid.UUID) error
	GetTags(ctx context.Context, todoID uuid.UUID) ([]domain.Tag, error)
	// Attachment associations (using simple string array)
	AddAttachment(ctx context.Context, todoID, userID uuid.UUID, attachmentID string) error
	RemoveAttachment(ctx context.Context, todoID, userID uuid.UUID, attachmentID string) error
	SetAttachments(ctx context.Context, todoID, userID uuid.UUID, attachmentIDs []string) error
}

type SubtaskRepository interface {
	Create(ctx context.Context, subtask *domain.Subtask) (*domain.Subtask, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Subtask, error)
	ListByTodo(ctx context.Context, todoID, userID uuid.UUID) ([]domain.Subtask, error)
	Update(ctx context.Context, id, userID uuid.UUID, updateData *domain.Subtask) (*domain.Subtask, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	GetParentTodoID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
}

// Transactioner interface allows services to run operations within a DB transaction
type Transactioner interface {
	BeginTx(ctx context.Context) (*db.Queries, error)
}

// RepositoryRegistry bundles all repositories together, often useful for dependency injection
type RepositoryRegistry struct {
	UserRepo    UserRepository
	TagRepo     TagRepository
	TodoRepo    TodoRepository
	SubtaskRepo SubtaskRepository
	*db.Queries
	Pool *pgxpool.Pool
}

// NewRepositoryRegistry creates a new registry
func NewRepositoryRegistry(pool *pgxpool.Pool) *RepositoryRegistry {
	queries := db.New(pool)
	return &RepositoryRegistry{
		UserRepo:    NewPgxUserRepository(queries),
		TagRepo:     NewPgxTagRepository(queries),
		TodoRepo:    NewPgxTodoRepository(queries, pool),
		SubtaskRepo: NewPgxSubtaskRepository(queries),
		Queries:     queries,
		Pool:        pool,
	}
}
