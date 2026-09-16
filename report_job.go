package jobs

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/codersirojiddin/reddit-leads/internal/keywords"
	"github.com/codersirojiddin/reddit-leads/internal/opportunities"
	"github.com/codersirojiddin/reddit-leads/internal/projects"
	"github.com/codersirojiddin/reddit-leads/internal/reports"
)

// ReportJob processes a single pending report: run discovery+scoring for the
// project's keywords, then rank the top opportunities into report_items.
type ReportJob struct {
	projectsRepo *projects.Repository
	keywordsRepo *keywords.Repository
	oppsSvc      *opportunities.Service
	reportsRepo  *reports.Repository
}

func NewReportJob(
	projectsRepo *projects.Repository,
	keywordsRepo *keywords.Repository,
	oppsSvc *opportunities.Service,
	reportsRepo *reports.Repository,
) *ReportJob {
	return &ReportJob{
		projectsRepo: projectsRepo,
		keywordsRepo: keywordsRepo,
		oppsSvc:      oppsSvc,
		reportsRepo:  reportsRepo,
	}
}

// Run executes the full pipeline for the given report and updates its status
// (completed/failed) accordingly. Any error is also stored on the report row
// so it's visible via the API.
func (j *ReportJob) Run(ctx context.Context, report *reports.Report) error {
	if err := j.process(ctx, report); err != nil {
		_ = j.reportsRepo.MarkFailed(ctx, report.ID, err.Error())
		return err
	}
	return j.reportsRepo.MarkCompleted(ctx, report.ID)
}

func (j *ReportJob) process(ctx context.Context, report *reports.Report) error {
	project, err := j.projectsRepo.FindByID(ctx, report.ProjectID)
	if err != nil {
		return fmt.Errorf("load project: %w", err)
	}

	kws, err := j.keywordsRepo.ListByProject(ctx, project.ID)
	if err != nil {
		return fmt.Errorf("load keywords: %w", err)
	}
	if len(kws) == 0 {
		return fmt.Errorf("project has no keywords to search")
	}

	inputs := make([]opportunities.KeywordInput, 0, len(kws))
	for _, k := range kws {
		inputs = append(inputs, opportunities.KeywordInput{ID: k.ID, Text: k.Keyword})
	}

	if _, err := j.oppsSvc.Run(ctx, project.ID, project.ProductDescription, inputs); err != nil {
		return fmt.Errorf("run discovery/scoring: %w", err)
	}

	top, err := j.oppsSvc.TopForProject(ctx, project.ID, reports.TopOpportunitiesPerReport)
	if err != nil {
		return fmt.Errorf("load top opportunities: %w", err)
	}

	oppIDs := make([]uuid.UUID, 0, len(top))
	for _, o := range top {
		oppIDs = append(oppIDs, o.ID)
	}

	if err := j.reportsRepo.AddItems(ctx, report.ID, oppIDs); err != nil {
		return fmt.Errorf("save report items: %w", err)
	}

	return nil
}
