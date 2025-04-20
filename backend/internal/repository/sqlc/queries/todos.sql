-- name: CreateTodo :one
INSERT INTO todos (user_id, title, description, status, deadline)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTodoByID :one
SELECT * FROM todos
WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: ListUserTodos :many
SELECT t.* FROM todos t
LEFT JOIN todo_tags tt ON t.id = tt.todo_id
WHERE
  t.user_id = sqlc.arg('user_id') -- Use sqlc.arg for required params
  AND (sqlc.narg('status_filter')::todo_status IS NULL OR t.status = sqlc.narg('status_filter'))
  AND (sqlc.narg('tag_id_filter')::uuid IS NULL OR tt.tag_id = sqlc.narg('tag_id_filter'))
  AND (sqlc.narg('deadline_before_filter')::timestamptz IS NULL OR t.deadline < sqlc.narg('deadline_before_filter'))
  AND (sqlc.narg('deadline_after_filter')::timestamptz IS NULL OR t.deadline > sqlc.narg('deadline_after_filter'))
GROUP BY t.id -- Still needed due to LEFT JOIN potentially multiplying rows if a todo has multiple tags
ORDER BY t.created_at DESC -- Or your desired order
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: UpdateTodo :one
UPDATE todos
SET
  title = COALESCE(sqlc.narg(title), title),
  description = sqlc.narg(description), -- Allow setting description to NULL
  status = COALESCE(sqlc.narg(status), status),
  deadline = sqlc.narg(deadline),       -- Allow setting deadline to NULL
  attachments = COALESCE(sqlc.narg(attachments), attachments)
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteTodo :exec
DELETE FROM todos
WHERE id = $1 AND user_id = $2;

-- name: AddAttachmentToTodo :exec
UPDATE todos
SET attachments = array_append(attachments, $1)
WHERE id = $2 AND user_id = $3;

-- name: RemoveAttachmentFromTodo :exec
UPDATE todos
SET attachments = array_remove(attachments, $1)
WHERE id = $2 AND user_id = $3;