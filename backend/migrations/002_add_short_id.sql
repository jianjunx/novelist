-- Add short_id column to projects table
ALTER TABLE projects ADD COLUMN short_id TEXT UNIQUE NOT NULL DEFAULT '';

-- Backfill existing projects with random short IDs
UPDATE projects SET short_id = encode(gen_random_bytes(4), 'hex') WHERE short_id = '';

-- Add NOT NULL constraint after backfill
ALTER TABLE projects ALTER COLUMN short_id DROP DEFAULT;
