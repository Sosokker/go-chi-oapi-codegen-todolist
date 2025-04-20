package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Sosokker/todolist-backend/internal/domain"
	db "github.com/Sosokker/todolist-backend/internal/repository/sqlc/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgxTodoRepository struct {
	q    *db.Queries
	pool *pgxpool.Pool
	// Consider adding a TagRepository dependency here for batch loading if needed
}

func NewPgxTodoRepository(queries *db.Queries, pool *pgxpool.Pool) TodoRepository {
	return &pgxTodoRepository{q: queries, pool: pool}
}

// --- Mapping functions ---

func mapDbTodoToDomain(dbTodo db.Todo) *domain.Todo {
	return &domain.Todo{
		ID:            dbTodo.ID,
		UserID:        dbTodo.UserID,
		Title:         dbTodo.Title,
		Description:   domain.NullStringToStringPtr(dbTodo.Description),
		Status:        domain.TodoStatus(dbTodo.Status),
		AttachmentUrl: domain.NullStringToStringPtr(dbTodo.AttachmentUrl),
		Deadline:      dbTodo.Deadline,
		CreatedAt:     dbTodo.CreatedAt,
		UpdatedAt:     dbTodo.UpdatedAt,
	}
}

func mapDbTagToDomain(dbTag db.Tag) domain.Tag {
	return domain.Tag{
		ID:        dbTag.ID,
		UserID:    dbTag.UserID,
		Name:      dbTag.Name,
		Color:     domain.NullStringToStringPtr(nullStringFromText(dbTag.Color)),
		Icon:      domain.NullStringToStringPtr(nullStringFromText(dbTag.Icon)),
		CreatedAt: dbTag.CreatedAt,
		UpdatedAt: dbTag.UpdatedAt,
	}
}

// ――― TodoRepository methods ―――

func (r *pgxTodoRepository) Create(
	ctx context.Context,
	todo *domain.Todo,
) (*domain.Todo, error) {
	params := db.CreateTodoParams{
		UserID:      todo.UserID,
		Title:       todo.Title,
		Description: sql.NullString{String: derefString(todo.Description), Valid: todo.Description != nil},
		Status:      db.TodoStatus(todo.Status),
		Deadline:    todo.Deadline,
	}
	dbTodo, err := r.q.CreateTodo(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}
	return mapDbTodoToDomain(dbTodo), nil
}

func (r *pgxTodoRepository) GetByID(
	ctx context.Context,
	id, userID uuid.UUID,
) (*domain.Todo, error) {
	dbTodo, err := r.q.GetTodoByID(ctx, db.GetTodoByIDParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}
	return mapDbTodoToDomain(dbTodo), nil
}

func (r *pgxTodoRepository) ListByUser(
	ctx context.Context,
	params ListTodosParams,
) ([]domain.Todo, error) {

	sqlcParams := db.ListUserTodosParams{
		UserID:               params.UserID,
		Limit:                int32(params.Limit),
		Offset:               int32(params.Offset),
		StatusFilter:         db.NullTodoStatus{Valid: false},
		TagIDFilter:          pgtype.UUID{Valid: false},
		DeadlineBeforeFilter: nil,
		DeadlineAfterFilter:  nil,
	}

	if params.Status != nil {
		sqlcParams.StatusFilter = db.NullTodoStatus{
			TodoStatus: db.TodoStatus(*params.Status),
			Valid:      true,
		}
	}

	if params.TagID != nil {
		sqlcParams.TagIDFilter = pgtype.UUID{
			Bytes: *params.TagID,
			Valid: true,
		}
	}

	if params.DeadlineBefore != nil {
		sqlcParams.DeadlineBeforeFilter = params.DeadlineBefore
	}

	if params.DeadlineAfter != nil {
		sqlcParams.DeadlineAfterFilter = params.DeadlineAfter
	}

	// Call the regenerated sqlc function with the correctly populated struct
	dbTodos, err := r.q.ListUserTodos(ctx, sqlcParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []domain.Todo{}, nil
		}
		return nil, fmt.Errorf("failed to list todos: %w", err)
	}

	todos := make([]domain.Todo, len(dbTodos))
	for i, t := range dbTodos {
		mappedTodo := mapDbTodoToDomain(t)
		if mappedTodo != nil {
			todos[i] = *mappedTodo
		} else {
			return nil, fmt.Errorf("failed to map database todo at index %d", i)
		}
	}
	return todos, nil
}

func (r *pgxTodoRepository) Update(
	ctx context.Context,
	id, userID uuid.UUID,
	updateData *domain.Todo,
) (*domain.Todo, error) {
	if _, err := r.GetByID(ctx, id, userID); err != nil {
		return nil, err
	}

	params := db.UpdateTodoParams{
		ID:          id,
		UserID:      userID,
		Title:       pgtype.Text{String: updateData.Title, Valid: true},
		Description: sql.NullString{String: derefString(updateData.Description), Valid: updateData.Description != nil},
		Status:      db.NullTodoStatus{TodoStatus: db.TodoStatus(updateData.Status), Valid: true},
		Deadline:    updateData.Deadline,
	}

	dbTodo, err := r.q.UpdateTodo(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, fmt.Errorf("foreign key violation: %w", domain.ErrBadRequest)
		}
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}
	return mapDbTodoToDomain(dbTodo), nil
}

func (r *pgxTodoRepository) Delete(
	ctx context.Context,
	id, userID uuid.UUID,
) error {
	if err := r.q.DeleteTodo(ctx, db.DeleteTodoParams{ID: id, UserID: userID}); err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}
	return nil
}

// --- Tag Associations ---

func (r *pgxTodoRepository) AddTag(
	ctx context.Context,
	todoID, tagID uuid.UUID,
) error {
	if err := r.q.AddTagToTodo(ctx, db.AddTagToTodoParams{TodoID: todoID, TagID: tagID}); err != nil {
		return fmt.Errorf("failed to add tag: %w", err)
	}
	return nil
}

func (r *pgxTodoRepository) RemoveTag(
	ctx context.Context,
	todoID, tagID uuid.UUID,
) error {
	if err := r.q.RemoveTagFromTodo(ctx, db.RemoveTagFromTodoParams{TodoID: todoID, TagID: tagID}); err != nil {
		return fmt.Errorf("failed to remove tag: %w", err)
	}
	return nil
}

func (r *pgxTodoRepository) SetTags(
	ctx context.Context,
	todoID uuid.UUID,
	tagIDs []uuid.UUID,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)
	if err := qtx.RemoveAllTagsFromTodo(ctx, todoID); err != nil {
		return fmt.Errorf("remove existing tags: %w", err)
	}
	for _, tID := range tagIDs {
		if err := qtx.AddTagToTodo(ctx, db.AddTagToTodoParams{TodoID: todoID, TagID: tID}); err != nil {
			return fmt.Errorf("add tag %s: %w", tID, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tags tx: %w", err)
	}
	return nil
}

func (r *pgxTodoRepository) GetTags(
	ctx context.Context,
	todoID uuid.UUID,
) ([]domain.Tag, error) {
	dbTags, err := r.q.GetTagsForTodo(ctx, todoID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []domain.Tag{}, nil
		}
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	tags := make([]domain.Tag, len(dbTags))
	for i, t := range dbTags {
		tags[i] = mapDbTagToDomain(t)
	}
	return tags, nil
}

func (r *pgxTodoRepository) UpdateAttachmentURL(
	ctx context.Context,
	todoID, userID uuid.UUID,
	attachmentURL *string,
) error {
	query := `
		UPDATE todos
		SET attachment_url = $1
		WHERE id = $2 AND user_id = $3
	`
	_, err := r.pool.Exec(ctx, query, attachmentURL, todoID, userID)
	if err != nil {
		return fmt.Errorf("failed to update attachment URL: %w", err)
	}
	return nil
}

// ― Helpers ―

func derefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
