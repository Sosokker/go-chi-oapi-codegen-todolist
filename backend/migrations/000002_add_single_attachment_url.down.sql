-- backend/migrations/000002_add_single_attachment_url.down.sql
-- Re-add the old array column and table (might lose data)
ALTER TABLE todos
ADD COLUMN attachments TEXT[] NOT NULL DEFAULT '{}';

CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    todo_id UUID NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    storage_path VARCHAR(512) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_attachments_todo_id ON attachments(todo_id);

-- Drop the new single URL column
ALTER TABLE todos
DROP COLUMN IF EXISTS attachment_url;