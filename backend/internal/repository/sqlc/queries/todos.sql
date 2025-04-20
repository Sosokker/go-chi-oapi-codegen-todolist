-- name: CreateTodo :one
INSERT INTO todos (user_id, title, description, status, deadline, attachment_url)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetTodoByID :one
SELECT * FROM todos
WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: ListUserTodos :many
SELECT t.* FROM todos t
LEFT JOIN todo_tags tt ON t.id = tt.todo_id
WHERE
  t.user_id = sqlc.arg('user_id')
  AND (sqlc.narg('status_filter')::todo_status IS NULL OR t.status = sqlc.narg('status_filter'))
  AND (sqlc.narg('tag_id_filter')::uuid IS NULL OR tt.tag_id = sqlc.narg('tag_id_filter'))
  AND (sqlc.narg('deadline_before_filter')::timestamptz IS NULL OR t.deadline < sqlc.narg('deadline_before_filter'))
  AND (sqlc.narg('deadline_after_filter')::timestamptz IS NULL OR t.deadline > sqlc.narg('deadline_after_filter'))
GROUP BY t.id
ORDER BY t.created_at DESC
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: UpdateTodo :one
UPDATE todos
SET
  title = COALESCE(sqlc.narg(title), title),
  description = sqlc.narg(description),
  status = COALESCE(sqlc.narg(status), status),
  deadline = sqlc.narg(deadline),
  attachment_url = COALESCE(sqlc.narg(attachment_url), attachment_url) -- Update attachment_url
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteTodo :exec
DELETE FROM todos
WHERE id = $1 AND user_id = $2;

-- name: UpdateTodoAttachmentURL :exec
-- Sets or clears the attachment URL for a specific todo
UPDATE todos
SET attachment_url = $1 -- $1 will be the URL (TEXT) or NULL
WHERE id = $2 AND user_id = $3;