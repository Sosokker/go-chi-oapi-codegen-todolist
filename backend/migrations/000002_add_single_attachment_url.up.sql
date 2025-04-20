-- backend/migrations/000002_add_single_attachment_url.up.sql
ALTER TABLE todos
ADD COLUMN attachment_url TEXT NULL;

-- Optional: Add a comment for clarity
COMMENT ON COLUMN todos.attachment_url IS 'Publicly accessible URL for the single image attachment';

-- Drop the old attachments array column and the separate attachments table
ALTER TABLE todos DROP COLUMN IF EXISTS attachments;
DROP TABLE IF EXISTS attachments; -- Cascade should handle FKs if any existed, but we assume it's clean

-- NOTE: No data migration from TEXT[] to TEXT is included here for simplicity.
-- In a real scenario, you might add logic here to migrate the first element
-- of the old array if it represented a URL, but that depends heavily on previous data.