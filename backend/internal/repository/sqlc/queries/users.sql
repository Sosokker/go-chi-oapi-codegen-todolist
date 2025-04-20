-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, email_verified, google_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByGoogleID :one
SELECT * FROM users
WHERE google_id = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET
  username = COALESCE(sqlc.narg(username), username),
  email = COALESCE(sqlc.narg(email), email),
  email_verified = COALESCE(sqlc.narg(email_verified), email_verified),
  google_id = COALESCE(sqlc.narg(google_id), google_id)
  -- password_hash update should be handled separately if needed
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;