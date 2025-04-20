-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users Table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL, -- Store hashed password
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    google_id VARCHAR(255) UNIQUE NULL, -- For Google OAuth association
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add index for faster email lookup
CREATE INDEX idx_users_email ON users(email);

-- Tags Table
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7) NULL, -- e.g., #FF5733
    icon VARCHAR(30) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Ensure tag name is unique per user
    UNIQUE (user_id, name)
);

-- Index for faster tag lookup by user
CREATE INDEX idx_tags_user_id ON tags(user_id);

-- Todos Table
CREATE TYPE todo_status AS ENUM ('pending', 'in-progress', 'completed');

CREATE TABLE todos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT NULL,
    status todo_status NOT NULL DEFAULT 'pending',
    deadline TIMESTAMPTZ NULL,
    -- attachments array will store file IDs or identifiers
    attachments TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common filtering/sorting
CREATE INDEX idx_todos_user_id ON todos(user_id);
CREATE INDEX idx_todos_status ON todos(status);
CREATE INDEX idx_todos_deadline ON todos(deadline);

-- Subtasks Table
CREATE TABLE subtasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    todo_id UUID NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for faster subtask lookup by todo
CREATE INDEX idx_subtasks_todo_id ON subtasks(todo_id);

-- Todo_Tags Junction Table (Many-to-Many)
CREATE TABLE todo_tags (
    todo_id UUID NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (todo_id, tag_id)
);

-- Index for faster lookup when filtering todos by tag
CREATE INDEX idx_todo_tags_tag_id ON todo_tags(tag_id);

-- Optional: Attachments Metadata Table (if storing more than just IDs/URLs in Todo)
-- Consider if you need detailed tracking, otherwise the TEXT[] on todos might suffice
CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    todo_id UUID NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- Redundant? Maybe for direct permission checks
    file_name VARCHAR(255) NOT NULL,
    storage_path VARCHAR(512) NOT NULL, -- e.g., S3 key or local path
    content_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_attachments_todo_id ON attachments(todo_id);


-- Function to automatically update 'updated_at' timestamp
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply the trigger to tables that have 'updated_at'
CREATE TRIGGER set_timestamp_users
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_tags
BEFORE UPDATE ON tags
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_todos
BEFORE UPDATE ON todos
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_subtasks
BEFORE UPDATE ON subtasks
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

-- Optional trigger for attachments table if created
CREATE TRIGGER set_timestamp_attachments
BEFORE UPDATE ON attachments
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();