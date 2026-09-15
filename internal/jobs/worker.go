package jobs

import (
	"context"
	"log"
	"time"
)

// Worker repeatedly polls the Queue for pending reports and runs ReportJob
// against each one it claims, until the context is cancelled.
type Worker struct {
	queue        *Queue
	job          *ReportJob
	pollInterval time.Duration
}

func NewWorker(queue *Queue, job *ReportJob, pollInterval time.Duration) *Worker {
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}
	return &Worker{queue: queue, job: job, pollInterval: pollInterval}
}

// Run blocks, polling for work until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	log.Printf("worker started, polling every %s", w.pollInterval)
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("worker shutting down")
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

// tick claims and processes as many pending reports as are available right
// now, then returns (the next tick will pick up more).
func (w *Worker) tick(ctx context.Context) {
	for {
		report, err := w.queue.ClaimNext(ctx)
		if err != nil {
			log.Printf("failed to claim next report: %v", err)
			return
		}
		if report == nil {
			return // no pending work
		}

		log.Printf("processing report %s (project %s)", report.ID, report.ProjectID)
		start := time.Now()
		if err := w.job.Run(ctx, report); err != nil {
			log.Printf("report %s failed after %s: %v", report.ID, time.Since(start), err)
			continue
		}
		log.Printf("report %s completed in %s", report.ID, time.Since(start))
	}
}
