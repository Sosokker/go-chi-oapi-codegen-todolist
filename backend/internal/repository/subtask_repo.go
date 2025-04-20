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
)

type pgxSubtaskRepository struct {
	q *db.Queries
}

func NewPgxSubtaskRepository(queries *db.Queries) SubtaskRepository {
	return &pgxSubtaskRepository{q: queries}
}

// --- Mapping functions ---
func mapDbSubtaskToDomain(d db.Subtask) *domain.Subtask {
	return &domain.Subtask{
		ID:          d.ID,
		TodoID:      d.TodoID,
		Description: d.Description,
		Completed:   d.Completed,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func mapDbSubtasksToDomain(ds []db.Subtask) []domain.Subtask {
	out := make([]domain.Subtask, len(ds))
	for i, d := range ds {
		out[i] = *mapDbSubtaskToDomain(d)
	}
	return out
}

// --- Repository Methods ---

func (r *pgxSubtaskRepository) Create(
	ctx context.Context,
	subtask *domain.Subtask,
) (*domain.Subtask, error) {
	params := db.CreateSubtaskParams{
		TodoID:      subtask.TodoID,
		Description: subtask.Description,
		Completed:   subtask.Completed,
	}
	d, err := r.q.CreateSubtask(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, fmt.Errorf("parent todo %s not found: %w", subtask.TodoID, domain.ErrBadRequest)
		}
		return nil, fmt.Errorf("failed to create subtask: %w", err)
	}
	return mapDbSubtaskToDomain(d), nil
}

func (r *pgxSubtaskRepository) GetByID(
	ctx context.Context,
	id, userID uuid.UUID,
) (*domain.Subtask, error) {
	d, err := r.q.GetSubtaskByID(ctx, db.GetSubtaskByIDParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get subtask: %w", err)
	}
	return mapDbSubtaskToDomain(d), nil
}

func (r *pgxSubtaskRepository) ListByTodo(
	ctx context.Context,
	todoID, userID uuid.UUID,
) ([]domain.Subtask, error) {
	ds, err := r.q.ListSubtasksForTodo(ctx, db.ListSubtasksForTodoParams{
		TodoID: todoID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []domain.Subtask{}, nil
		}
		return nil, fmt.Errorf("failed to list subtasks: %w", err)
	}
	return mapDbSubtasksToDomain(ds), nil
}

func (r *pgxSubtaskRepository) Update(
	ctx context.Context,
	id, userID uuid.UUID,
	updateData *domain.Subtask,
) (*domain.Subtask, error) {
	params := db.UpdateSubtaskParams{
		ID:     id,
		UserID: userID,
		Description: sql.NullString{
			String: updateData.Description,
			Valid:  updateData.Description != "",
		},
		Completed: pgtype.Bool{
			Bool:  updateData.Completed,
			Valid: true,
		},
	}

	d, err := r.q.UpdateSubtask(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to update subtask: %w", err)
	}
	return mapDbSubtaskToDomain(d), nil
}

func (r *pgxSubtaskRepository) Delete(
	ctx context.Context,
	id, userID uuid.UUID,
) error {
	if err := r.q.DeleteSubtask(ctx, db.DeleteSubtaskParams{
		ID:     id,
		UserID: userID,
	}); err != nil {
		return fmt.Errorf("failed to delete subtask: %w", err)
	}
	return nil
}

func (r *pgxSubtaskRepository) GetParentTodoID(
	ctx context.Context,
	id uuid.UUID,
) (uuid.UUID, error) {
	todoID, err := r.q.GetTodoIDForSubtask(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, domain.ErrNotFound
		}
		return uuid.Nil, fmt.Errorf("failed to get parent todo id: %w", err)
	}
	return todoID, nil
}
