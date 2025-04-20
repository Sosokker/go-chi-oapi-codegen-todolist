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

type pgxTagRepository struct {
	q *db.Queries
}

func NewPgxTagRepository(queries *db.Queries) TagRepository {
	return &pgxTagRepository{q: queries}
}

func pgTextFromPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func nullStringFromText(t pgtype.Text) sql.NullString {
	return sql.NullString{String: t.String, Valid: t.Valid}
}

func mapDbTagToDomainTag(dbTag db.Tag) *domain.Tag {
	return &domain.Tag{
		ID:        dbTag.ID,
		UserID:    dbTag.UserID,
		Name:      dbTag.Name,
		Color:     domain.NullStringToStringPtr(nullStringFromText(dbTag.Color)),
		Icon:      domain.NullStringToStringPtr(nullStringFromText(dbTag.Icon)),
		CreatedAt: dbTag.CreatedAt,
		UpdatedAt: dbTag.UpdatedAt,
	}
}

func mapDbTagsToDomainTags(dbTags []db.Tag) []domain.Tag {
	tags := make([]domain.Tag, len(dbTags))
	for i, t := range dbTags {
		tags[i] = *mapDbTagToDomainTag(t)
	}
	return tags
}

func (r *pgxTagRepository) Create(
	ctx context.Context,
	tag *domain.Tag,
) (*domain.Tag, error) {
	params := db.CreateTagParams{
		UserID: tag.UserID,
		Name:   tag.Name,
		Color:  pgTextFromPtr(tag.Color),
		Icon:   pgTextFromPtr(tag.Icon),
	}

	dbTag, err := r.q.CreateTag(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("tag name '%s' already exists: %w", tag.Name, domain.ErrConflict)
		}
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}
	return mapDbTagToDomainTag(dbTag), nil
}

func (r *pgxTagRepository) GetByID(
	ctx context.Context,
	id, userID uuid.UUID,
) (*domain.Tag, error) {
	dbTag, err := r.q.GetTagByID(ctx, db.GetTagByIDParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get tag by id: %w", err)
	}
	return mapDbTagToDomainTag(dbTag), nil
}

func (r *pgxTagRepository) GetByIDs(
	ctx context.Context,
	ids []uuid.UUID,
	userID uuid.UUID,
) ([]domain.Tag, error) {
	if len(ids) == 0 {
		return []domain.Tag{}, nil
	}
	dbTags, err := r.q.GetTagsByIDs(ctx, db.GetTagsByIDsParams{
		UserID: userID,
		TagIds: ids,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []domain.Tag{}, nil
		}
		return nil, fmt.Errorf("failed to get tags by ids: %w", err)
	}
	return mapDbTagsToDomainTags(dbTags), nil
}

func (r *pgxTagRepository) ListByUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]domain.Tag, error) {
	dbTags, err := r.q.ListUserTags(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []domain.Tag{}, nil
		}
		return nil, fmt.Errorf("failed to list user tags: %w", err)
	}
	return mapDbTagsToDomainTags(dbTags), nil
}

func (r *pgxTagRepository) Update(
	ctx context.Context,
	id, userID uuid.UUID,
	updateData *domain.Tag,
) (*domain.Tag, error) {
	if _, err := r.GetByID(ctx, id, userID); err != nil {
		return nil, err
	}

	params := db.UpdateTagParams{
		ID:     id,
		UserID: userID,
		Name:   pgtype.Text{String: updateData.Name, Valid: true},
		Color:  pgTextFromPtr(updateData.Color),
		Icon:   pgTextFromPtr(updateData.Icon),
	}

	dbTag, err := r.q.UpdateTag(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("tag name '%s' already exists: %w", updateData.Name, domain.ErrConflict)
		}
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}
	return mapDbTagToDomainTag(dbTag), nil
}

func (r *pgxTagRepository) Delete(
	ctx context.Context,
	id, userID uuid.UUID,
) error {
	if err := r.q.DeleteTag(ctx, db.DeleteTagParams{ID: id, UserID: userID}); err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}
	return nil
}
