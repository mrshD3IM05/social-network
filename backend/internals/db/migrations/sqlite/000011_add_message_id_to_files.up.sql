ALTER TABLE files ADD COLUMN message_id INTEGER REFERENCES messages(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_files_message ON files(message_id);
