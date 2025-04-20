-- name: AddTagToTodo :exec
INSERT INTO todo_tags (todo_id, tag_id)
VALUES ($1, $2)
ON CONFLICT (todo_id, tag_id) DO NOTHING; -- Ignore if already exists

-- name: RemoveTagFromTodo :exec
DELETE FROM todo_tags
WHERE todo_id = $1 AND tag_id = $2;

-- name: RemoveAllTagsFromTodo :exec
DELETE FROM todo_tags
WHERE todo_id = $1;

-- name: GetTagsForTodo :many
SELECT t.*
FROM tags t
JOIN todo_tags tt ON t.id = tt.tag_id
WHERE tt.todo_id = $1;