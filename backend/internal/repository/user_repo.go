package repository

import (
	"context"
	"errors"

	"github.com/Sosokker/todolist-backend/internal/domain"
	db "github.com/Sosokker/todolist-backend/internal/repository/sqlc/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type pgxUserRepository struct {
	q *db.Queries
}

func NewPgxUserRepository(queries *db.Queries) UserRepository {
	return &pgxUserRepository{q: queries}
}

// mapDbUserToDomain converts a generated User → domain.User
func mapDbUserToDomain(u db.User) *domain.User {
	var googleID *string
	if u.GoogleID.Valid {
		googleID = &u.GoogleID.String
	}
	return &domain.User{
		ID:            u.ID,
		Username:      u.Username,
		Email:         u.Email,
		PasswordHash:  u.PasswordHash,
		EmailVerified: u.EmailVerified,
		GoogleID:      googleID,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

func (r *pgxUserRepository) Create(
	ctx context.Context,
	user *domain.User,
) (*domain.User, error) {
	var pgGoogleID pgtype.Text
	if user.GoogleID != nil {
		pgGoogleID = pgtype.Text{String: *user.GoogleID, Valid: true}
	}

	dbUser, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Username:      user.Username,
		Email:         user.Email,
		PasswordHash:  user.PasswordHash,
		EmailVerified: user.EmailVerified,
		GoogleID:      pgGoogleID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrConflict
		}
		return nil, err
	}
	return mapDbUserToDomain(dbUser), nil
}

func (r *pgxUserRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.User, error) {
	dbUser, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return mapDbUserToDomain(dbUser), nil
}

func (r *pgxUserRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	dbUser, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return mapDbUserToDomain(dbUser), nil
}

func (r *pgxUserRepository) GetByGoogleID(
	ctx context.Context,
	googleID string,
) (*domain.User, error) {
	pgGoogleID := pgtype.Text{String: googleID, Valid: true}
	dbUser, err := r.q.GetUserByGoogleID(ctx, pgGoogleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return mapDbUserToDomain(dbUser), nil
}

func (r *pgxUserRepository) Update(
	ctx context.Context,
	id uuid.UUID,
	u *domain.User,
) (*domain.User, error) {
	if _, err := r.GetByID(ctx, id); err != nil {
		return nil, err
	}

	var username pgtype.Text
	if u.Username != "" {
		username = pgtype.Text{String: u.Username, Valid: true}
	}
	var email pgtype.Text
	if u.Email != "" {
		email = pgtype.Text{String: u.Email, Valid: true}
	}
	emailVerified := pgtype.Bool{Bool: u.EmailVerified, Valid: true}

	var pgGoogleID pgtype.Text
	if u.GoogleID != nil {
		pgGoogleID = pgtype.Text{String: *u.GoogleID, Valid: true}
	}

	dbUser, err := r.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:            id,
		Username:      username,
		Email:         email,
		EmailVerified: emailVerified,
		GoogleID:      pgGoogleID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrConflict
		}
		return nil, err
	}
	return mapDbUserToDomain(dbUser), nil
}

func (r *pgxUserRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	return r.q.DeleteUser(ctx, id)
}
