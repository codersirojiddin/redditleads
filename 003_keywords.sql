CREATE TABLE IF NOT EXISTS keywords (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    keyword    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, keyword)
);

CREATE INDEX IF NOT EXISTS idx_keywords_project_id ON keywords(project_id);
