package reports

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("report not found")

const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

type Report struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	Status      string
	Error       string
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
}

type ReportItem struct {
	ID            uuid.UUID
	ReportID      uuid.UUID
	OpportunityID uuid.UUID
	Rank          int
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, projectID uuid.UUID) (*Report, error) {
	rep := &Report{}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO reports (project_id, status)
		VALUES ($1, $2)
		RETURNING id, project_id, status, error, created_at, started_at, completed_at
	`, projectID, StatusPending).Scan(
		&rep.ID, &rep.ProjectID, &rep.Status, &rep.Error, &rep.CreatedAt, &rep.StartedAt, &rep.CompletedAt,
	)
	if err != nil {
		return nil, err
	}
	return rep, nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*Report, error) {
	rep := &Report{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, project_id, status, error, created_at, started_at, completed_at
		FROM reports WHERE id = $1
	`, id).Scan(&rep.ID, &rep.ProjectID, &rep.Status, &rep.Error, &rep.CreatedAt, &rep.StartedAt, &rep.CompletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return rep, nil
}

func (r *Repository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*Report, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, project_id, status, error, created_at, started_at, completed_at
		FROM reports WHERE project_id = $1 ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Report
	for rows.Next() {
		rep := &Report{}
		if err := rows.Scan(&rep.ID, &rep.ProjectID, &rep.Status, &rep.Error, &rep.CreatedAt, &rep.StartedAt, &rep.CompletedAt); err != nil {
			return nil, err
		}
		out = append(out, rep)
	}
	return out, rows.Err()
}

// ClaimNextPending atomically picks the oldest pending report and marks it
// running, so multiple worker instances don't process the same report twice.
func (r *Repository) ClaimNextPending(ctx context.Context) (*Report, error) {
	rep := &Report{}
	err := r.pool.QueryRow(ctx, `
		UPDATE reports
		SET status = $1, started_at = now()
		WHERE id = (
			SELECT id FROM reports
			WHERE status = $2
			ORDER BY created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, project_id, status, error, created_at, started_at, completed_at
	`, StatusRunning, StatusPending).Scan(
		&rep.ID, &rep.ProjectID, &rep.Status, &rep.Error, &rep.CreatedAt, &rep.StartedAt, &rep.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // no pending work, not an error
		}
		return nil, err
	}
	return rep, nil
}

func (r *Repository) MarkCompleted(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE reports SET status = $1, completed_at = now() WHERE id = $2
	`, StatusCompleted, id)
	return err
}

func (r *Repository) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE reports SET status = $1, error = $2, completed_at = now() WHERE id = $3
	`, StatusFailed, errMsg, id)
	return err
}

func (r *Repository) AddItems(ctx context.Context, reportID uuid.UUID, opportunityIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, oppID := range opportunityIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO report_items (report_id, opportunity_id, rank)
			VALUES ($1, $2, $3)
			ON CONFLICT (report_id, opportunity_id) DO NOTHING
		`, reportID, oppID, i+1); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *Repository) ItemsByReport(ctx context.Context, reportID uuid.UUID) ([]*ReportItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, report_id, opportunity_id, rank
		FROM report_items WHERE report_id = $1 ORDER BY rank ASC
	`, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*ReportItem
	for rows.Next() {
		it := &ReportItem{}
		if err := rows.Scan(&it.ID, &it.ReportID, &it.OpportunityID, &it.Rank); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}
