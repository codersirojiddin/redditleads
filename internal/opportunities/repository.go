package opportunities

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/codersirojiddin/reddit-leads/internal/discovery"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// UpsertRedditPost inserts a Reddit post for a project/keyword, or returns
// the existing row's ID if it was already fetched before (deduplication by
// project_id + reddit_id).
func (r *Repository) UpsertRedditPost(ctx context.Context, projectID, keywordID uuid.UUID, p discovery.RedditPost) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO reddit_posts
			(project_id, keyword_id, reddit_id, title, body, subreddit, author, url, reddit_score, num_comments, created_utc)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (project_id, reddit_id) DO UPDATE
			SET reddit_score = EXCLUDED.reddit_score,
			    num_comments = EXCLUDED.num_comments
		RETURNING id
	`,
		projectID, keywordID, p.ID, p.Title, p.Body, p.Subreddit, p.Author, p.URL,
		p.Score, p.NumComments, p.CreatedAt,
	).Scan(&id)
	return id, err
}

// OpportunityExists checks whether we've already scored this post for this
// project (so we don't re-spend an LLM call on repeat discovery runs).
func (r *Repository) OpportunityExists(ctx context.Context, projectID, redditPostID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM opportunities WHERE project_id = $1 AND reddit_post_id = $2)
	`, projectID, redditPostID).Scan(&exists)
	return exists, err
}

func (r *Repository) Create(ctx context.Context, o *Opportunity) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO opportunities
			(project_id, reddit_post_id, relevance_score, intent_score, rule_score, total_score, reasoning, suggested_reply)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (project_id, reddit_post_id) DO UPDATE
			SET relevance_score = EXCLUDED.relevance_score,
			    intent_score = EXCLUDED.intent_score,
			    rule_score = EXCLUDED.rule_score,
			    total_score = EXCLUDED.total_score,
			    reasoning = EXCLUDED.reasoning,
			    suggested_reply = EXCLUDED.suggested_reply
		RETURNING id
	`, o.ProjectID, o.RedditPostID, o.RelevanceScore, o.IntentScore, o.RuleScore, o.TotalScore, o.Reasoning, o.SuggestedReply,
	).Scan(&id)
	return id, err
}

// TopByProject returns the highest-scoring opportunities for a project,
// joined with their source post, ordered by total_score descending.
func (r *Repository) TopByProject(ctx context.Context, projectID uuid.UUID, limit int) ([]*Opportunity, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT o.id, o.project_id, o.reddit_post_id, o.relevance_score, o.intent_score,
		       o.rule_score, o.total_score, o.reasoning, o.suggested_reply, o.status, o.created_at,
		       rp.title, rp.url, rp.subreddit, rp.author
		FROM opportunities o
		JOIN reddit_posts rp ON rp.id = o.reddit_post_id
		WHERE o.project_id = $1
		ORDER BY o.total_score DESC
		LIMIT $2
	`, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Opportunity
	for rows.Next() {
		o := &Opportunity{}
		if err := rows.Scan(
			&o.ID, &o.ProjectID, &o.RedditPostID, &o.RelevanceScore, &o.IntentScore,
			&o.RuleScore, &o.TotalScore, &o.Reasoning, &o.SuggestedReply, &o.Status, &o.CreatedAt,
			&o.PostTitle, &o.PostURL, &o.PostSubreddit, &o.PostAuthor,
		); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// ByIDs fetches specific opportunities (joined with post info), preserving
// no particular order — callers should re-order by the ids slice if needed.
func (r *Repository) ByIDs(ctx context.Context, ids []uuid.UUID) ([]*Opportunity, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT o.id, o.project_id, o.reddit_post_id, o.relevance_score, o.intent_score,
		       o.rule_score, o.total_score, o.reasoning, o.suggested_reply, o.status, o.created_at,
		       rp.title, rp.url, rp.subreddit, rp.author
		FROM opportunities o
		JOIN reddit_posts rp ON rp.id = o.reddit_post_id
		WHERE o.id = ANY($1)
	`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Opportunity
	for rows.Next() {
		o := &Opportunity{}
		if err := rows.Scan(
			&o.ID, &o.ProjectID, &o.RedditPostID, &o.RelevanceScore, &o.IntentScore,
			&o.RuleScore, &o.TotalScore, &o.Reasoning, &o.SuggestedReply, &o.Status, &o.CreatedAt,
			&o.PostTitle, &o.PostURL, &o.PostSubreddit, &o.PostAuthor,
		); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}
