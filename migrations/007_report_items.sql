CREATE TABLE IF NOT EXISTS report_items (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id      UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    rank           INT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (report_id, opportunity_id)
);

CREATE INDEX IF NOT EXISTS idx_report_items_report_id ON report_items(report_id);
