-- Drop triggers first
DROP TRIGGER IF EXISTS set_timestamp_users ON users;
DROP TRIGGER IF EXISTS set_timestamp_tags ON tags;
DROP TRIGGER IF EXISTS set_timestamp_todos ON todos;
DROP TRIGGER IF EXISTS set_timestamp_subtasks ON subtasks;
DROP TRIGGER IF EXISTS set_timestamp_attachments ON attachments;

-- Drop the trigger function
DROP FUNCTION IF EXISTS trigger_set_timestamp();

-- Drop tables in reverse order of creation (or based on dependencies)
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS todo_tags;
DROP TABLE IF EXISTS subtasks;
DROP TABLE IF EXISTS todos;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS users;

-- Drop custom types
DROP TYPE IF EXISTS todo_status;

-- Drop extensions (usually not needed in down migration unless specifically required)
DROP EXTENSION IF EXISTS "uuid-ossp";