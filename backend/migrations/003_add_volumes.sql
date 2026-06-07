-- 003: Add volumes table (篇) between projects and outlines

CREATE TABLE IF NOT EXISTS volumes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    volume_num INT NOT NULL DEFAULT 1,
    title TEXT DEFAULT '',
    description TEXT DEFAULT '',
    status TEXT DEFAULT 'draft',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_volumes_project_id ON volumes(project_id);

-- Add volume_id to outlines
ALTER TABLE outlines ADD COLUMN IF NOT EXISTS volume_id UUID REFERENCES volumes(id);

CREATE INDEX IF NOT EXISTS idx_outlines_volume_id ON outlines(volume_id);

-- Backfill: create default volume for existing projects that have outlines
INSERT INTO volumes (project_id, volume_num, title)
SELECT DISTINCT o.project_id, 1, '第一篇'
FROM outlines o
WHERE NOT EXISTS (
    SELECT 1 FROM volumes v WHERE v.project_id = o.project_id AND v.volume_num = 1
);

-- Assign orphaned outlines to their project's default volume
UPDATE outlines SET volume_id = (
    SELECT v.id FROM volumes v
    WHERE v.project_id = outlines.project_id AND v.volume_num = 1
)
WHERE volume_id IS NULL;
