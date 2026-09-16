package keywords

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDuplicate = errors.New("keyword already exists for this project")

type Keyword struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Keyword   string
	CreatedAt time.Time
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CountByProject(ctx context.Context, projectID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM keywords WHERE project_id = $1`, projectID,
	).Scan(&n)
	return n, err
}

func (r *Repository) Create(ctx context.Context, projectID uuid.UUID, keyword string) (*Keyword, error) {
	k := &Keyword{}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO keywords (project_id, keyword)
		VALUES ($1, $2)
		RETURNING id, project_id, keyword, created_at
	`, projectID, keyword).Scan(&k.ID, &k.ProjectID, &k.Keyword, &k.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	return k, nil
}

func (r *Repository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*Keyword, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, project_id, keyword, created_at
		FROM keywords
		WHERE project_id = $1
		ORDER BY created_at ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Keyword
	for rows.Next() {
		k := &Keyword{}
		if err := rows.Scan(&k.ID, &k.ProjectID, &k.Keyword, &k.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM keywords WHERE id = $1`, id)
	return err
}

func isUniqueViolation(err error) bool {
	type pgErr interface{ SQLState() string }
	var pe pgErr
	if errors.As(err, &pe) {
		return pe.SQLState() == "23505"
	}
	return false
}
