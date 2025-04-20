-- name: CreateSubtask :one
INSERT INTO subtasks (todo_id, description, completed)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSubtaskByID :one
-- We need to join to check ownership via the parent todo
SELECT s.* FROM subtasks s
JOIN todos t ON s.todo_id = t.id
WHERE s.id = $1 AND t.user_id = $2 LIMIT 1;

-- name: ListSubtasksForTodo :many
SELECT s.* FROM subtasks s
JOIN todos t ON s.todo_id = t.id
WHERE s.todo_id = $1 AND t.user_id = $2
ORDER BY s.created_at ASC;

-- name: UpdateSubtask :one
-- Need to join to check ownership before updating
UPDATE subtasks s
SET
  description = COALESCE(sqlc.narg(description), s.description),
  completed = COALESCE(sqlc.narg(completed), s.completed)
FROM todos t -- Include todos table in FROM clause for WHERE condition
WHERE s.id = $1 AND s.todo_id = t.id AND t.user_id = $2
RETURNING s.*; -- Return columns from subtasks (aliased as s)

-- name: DeleteSubtask :exec
-- Need owner check before deleting
DELETE FROM subtasks s
USING todos t
WHERE s.id = $1 AND s.todo_id = t.id AND t.user_id = $2;

-- name: GetTodoIDForSubtask :one
-- Helper to get parent todo ID for authorization checks in service layer if needed
SELECT todo_id FROM subtasks WHERE id = $1;