CREATE TABLE IF NOT EXISTS reddit_posts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id     UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    keyword_id     UUID NOT NULL REFERENCES keywords(id) ON DELETE CASCADE,
    reddit_id      TEXT NOT NULL,
    title          TEXT NOT NULL,
    body           TEXT NOT NULL DEFAULT '',
    subreddit      TEXT NOT NULL,
    author         TEXT NOT NULL DEFAULT '',
    url            TEXT NOT NULL,
    reddit_score   INT NOT NULL DEFAULT 0,
    num_comments   INT NOT NULL DEFAULT 0,
    created_utc    BIGINT NOT NULL,
    fetched_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, reddit_id)
);

CREATE INDEX IF NOT EXISTS idx_reddit_posts_project_id ON reddit_posts(project_id);
