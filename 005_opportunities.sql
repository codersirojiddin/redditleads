CREATE TABLE IF NOT EXISTS opportunities (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id       UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    reddit_post_id   UUID NOT NULL REFERENCES reddit_posts(id) ON DELETE CASCADE,
    relevance_score  NUMERIC(5,2) NOT NULL DEFAULT 0,
    intent_score     NUMERIC(5,2) NOT NULL DEFAULT 0,
    rule_score       NUMERIC(5,2) NOT NULL DEFAULT 0,
    total_score      NUMERIC(5,2) NOT NULL DEFAULT 0,
    reasoning        TEXT NOT NULL DEFAULT '',
    suggested_reply  TEXT NOT NULL DEFAULT '',
    status           TEXT NOT NULL DEFAULT 'new',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, reddit_post_id)
);

CREATE INDEX IF NOT EXISTS idx_opportunities_project_id ON opportunities(project_id);
CREATE INDEX IF NOT EXISTS idx_opportunities_total_score ON opportunities(total_score DESC);
