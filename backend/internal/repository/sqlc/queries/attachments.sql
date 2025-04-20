-- name: CreateAttachment :one
INSERT INTO attachments (todo_id, user_id, file_name, storage_path, content_type, size)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetAttachmentByID :one
SELECT * FROM attachments
WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: ListAttachmentsForTodo :many
SELECT * FROM attachments
WHERE todo_id = $1 AND user_id = $2
ORDER BY uploaded_at ASC;

-- name: DeleteAttachment :exec
DELETE FROM attachments
WHERE id = $1 AND user_id = $2;