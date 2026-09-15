package projects

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("project not found")

type Project struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	Name               string
	ProductDescription string
	TargetURL          string
	CreatedAt          time.Time
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, p *Project) (*Project, error) {
	out := &Project{}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO projects (user_id, name, product_description, target_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, name, product_description, COALESCE(target_url, ''), created_at
	`, p.UserID, p.Name, p.ProductDescription, p.TargetURL).Scan(
		&out.ID, &out.UserID, &out.Name, &out.ProductDescription, &out.TargetURL, &out.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*Project, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, product_description, COALESCE(target_url, ''), created_at
		FROM projects
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Project
	for rows.Next() {
		p := &Project{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.ProductDescription, &p.TargetURL, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*Project, error) {
	p := &Project{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, product_description, COALESCE(target_url, ''), created_at
		FROM projects WHERE id = $1
	`, id).Scan(&p.ID, &p.UserID, &p.Name, &p.ProductDescription, &p.TargetURL, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
