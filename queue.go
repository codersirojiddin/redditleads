package jobs

import (
	"context"

	"github.com/codersirojiddin/reddit-leads/internal/reports"
)

// Queue is a thin abstraction over the reports table used as a simple
// DB-backed job queue: a "job" is just a report row in status=pending.
// This avoids needing a separate broker (Redis/SQS) for the MVP.
type Queue struct {
	reportsRepo *reports.Repository
}

func NewQueue(reportsRepo *reports.Repository) *Queue {
	return &Queue{reportsRepo: reportsRepo}
}

// ClaimNext atomically claims the oldest pending report, if any. Returns nil
// (no error) if there's currently no pending work.
func (q *Queue) ClaimNext(ctx context.Context) (*reports.Report, error) {
	return q.reportsRepo.ClaimNextPending(ctx)
}
