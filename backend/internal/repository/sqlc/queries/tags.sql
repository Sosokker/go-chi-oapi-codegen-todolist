-- name: CreateTag :one
INSERT INTO tags (user_id, name, color, icon)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetTagByID :one
SELECT * FROM tags
WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: ListUserTags :many
SELECT * FROM tags
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateTag :one
UPDATE tags
SET
  name = COALESCE(sqlc.narg(name), name),
  color = sqlc.narg(color), -- Allow setting color to NULL
  icon = sqlc.narg(icon)   -- Allow setting icon to NULL
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteTag :exec
DELETE FROM tags
WHERE id = $1 AND user_id = $2;

-- name: GetTagsByIDs :many
SELECT * FROM tags
WHERE id = ANY(sqlc.arg(tag_ids)::uuid[]) AND user_id = $1;